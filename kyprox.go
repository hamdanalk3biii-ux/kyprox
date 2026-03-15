package main

import (
	"bufio"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"golang.org/x/net/proxy"
)

/*
Embed ads.json into binary
*/

//go:embed ads.json
var adsData []byte

type ProxyInfo struct {
	Proxy       string `json:"proxy"`
	Latency     int    `json:"latency"`
	Country     string `json:"country"`
	City        string `json:"city"`
	LastChecked int64  `json:"last_checked"`
}

type ProxyResponse struct {
	Protocol string      `json:"protocol"`
	Count    int         `json:"count"`
	Proxies  []ProxyInfo `json:"proxies"`
}

var blockedDomains []string
var adsEnabled bool

func main() {

	loadAds()

	reader := bufio.NewReader(os.Stdin)

	printBanner()

	for {

		printMenu()

		fmt.Print("Select option: ")

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {

		case "1":
			fetchAndConnectProxy(reader)

		case "2":
			manualProxy(reader)

		case "3":

			adsEnabled = !adsEnabled

			if adsEnabled {
				color.Green("AdBlock ENABLED")
			} else {
				color.Red("AdBlock DISABLED")
			}

		case "4":
			showHelp()

		case "5":
			os.Exit(0)

		default:
			color.Red("Invalid option")
		}
	}
}

func printBanner() {

	color.Cyan("╔══════════════════════════════════════╗")
	color.Cyan("║              KYPROX                  ║")
	color.Cyan("║        Termux Proxy Manager          ║")
	color.Cyan("╚══════════════════════════════════════╝")
}

func printMenu() {

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════╗")
	fmt.Println("║ 1  Fetch Proxy From API              ║")
	fmt.Println("║ 2  Enter Manual Proxy                ║")
	fmt.Println("║ 3  Toggle AdBlock                    ║")
	fmt.Println("║ 4  Help                              ║")
	fmt.Println("║ 5  Exit                              ║")
	fmt.Println("╚══════════════════════════════════════╝")

	if adsEnabled {
		color.Green("AdBlock: ENABLED")
	} else {
		color.Red("AdBlock: DISABLED")
	}

	fmt.Println()
}

func loadAds() {

	err := json.Unmarshal(adsData, &blockedDomains)

	if err != nil {
		fmt.Println("Failed to load ad list")
		return
	}

	color.Green("Loaded %d ad domains", len(blockedDomains))
}

func isBlocked(host string) bool {

	if !adsEnabled {
		return false
	}

	for _, d := range blockedDomains {

		if strings.Contains(host, d) {
			return true
		}
	}

	return false
}

func fetchAndConnectProxy(reader *bufio.Reader) {

	resp, err := http.Get("https://libhub.stackverify.site/api/proxy.php?protocol=socks5&limit=5")

	if err != nil {
		log.Println("API fetch failed:", err)
		return
	}

	defer resp.Body.Close()

	var data ProxyResponse

	err = json.NewDecoder(resp.Body).Decode(&data)

	if err != nil {
		log.Println("JSON error:", err)
		return
	}

	printProxyTable(data.Proxies)

	fmt.Print("Select proxy number: ")

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	index, _ := strconv.Atoi(input)

	if index < 1 || index > len(data.Proxies) {

		color.Red("Invalid selection")
		return
	}

	selected := data.Proxies[index-1]

	startLocalProxy(selected.Proxy, "socks5", selected.Country, selected.City)
}

func printProxyTable(proxies []ProxyInfo) {

	fmt.Println()

	fmt.Println("╔════╦════════════════════╦════════════╦════════════╦════════╗")
	fmt.Println("║ #  ║ Proxy              ║ Country    ║ City       ║Latency ║")
	fmt.Println("╠════╬════════════════════╬════════════╬════════════╬════════╣")

	for i, p := range proxies {

		fmt.Printf("║ %-2d ║ %-18s ║ %-10s ║ %-10s ║ %-6d║\n",
			i+1,
			p.Proxy,
			p.Country,
			p.City,
			p.Latency)
	}

	fmt.Println("╚════╩════════════════════╩════════════╩════════════╩════════╝")
}

func manualProxy(reader *bufio.Reader) {

	fmt.Print("Enter proxy IP:PORT : ")

	proxyAddr, _ := reader.ReadString('\n')
	proxyAddr = strings.TrimSpace(proxyAddr)

	startLocalProxy(proxyAddr, "socks5", "Unknown", "Unknown")
}

func startLocalProxy(proxyAddr, proxyType, country, city string) {

	color.Green("╔══════════════════════════════════╗")
	color.Green("║          PROXY CONNECTED         ║")
	color.Green("╠══════════════════════════════════╣")

	fmt.Printf("║ Remote Proxy : %-16s║\n", proxyAddr)
	fmt.Printf("║ Location     : %-16s║\n", country+" "+city)
	fmt.Printf("║ Type         : %-16s║\n", proxyType)

	if adsEnabled {
		fmt.Printf("║ AdBlock      : ENABLED           ║\n")
	} else {
		fmt.Printf("║ AdBlock      : DISABLED          ║\n")
	}

	color.Green("╚══════════════════════════════════╝")

	dialer, err := proxy.SOCKS5("tcp", proxyAddr, nil, proxy.Direct)

	if err != nil {
		log.Fatal(err)
	}

	server := &http.Server{

		Addr: "0.0.0.0:8080",

		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if isBlocked(r.Host) {

				http.Error(w, "Blocked by AdBlock", http.StatusForbidden)
				return
			}

			if r.Method == http.MethodConnect {

				handleTunneling(w, r, dialer)

			} else {

				handleHTTP(w, r, dialer)
			}
		}),
	}

	color.Cyan("\nLocal Proxy Running")
	fmt.Println("Host : 127.0.0.1")
	fmt.Println("Port : 8080")

	fmt.Println("\nPress CTRL+C to stop")

	log.Fatal(server.ListenAndServe())
}

func handleTunneling(w http.ResponseWriter, r *http.Request, dialer proxy.Dialer) {

	destConn, err := dialer.Dial("tcp", r.Host)

	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	hijacker, _ := w.(http.Hijacker)

	clientConn, _, _ := hijacker.Hijack()

	fmt.Fprint(clientConn, "HTTP/1.1 200 Connection Established\r\n\r\n")

	go transfer(destConn, clientConn)
	go transfer(clientConn, destConn)
}

func handleHTTP(w http.ResponseWriter, r *http.Request, dialer proxy.Dialer) {

	req := r.Clone(r.Context())
	req.RequestURI = ""

	removeHeaders := []string{
		"X-Forwarded-For",
		"X-Real-IP",
		"Forwarded",
		"Via",
		"Client-IP",
		"CF-Connecting-IP",
		"True-Client-IP",
	}

	for _, h := range removeHeaders {
		req.Header.Del(h)
	}

	client := &http.Client{

		Transport: &http.Transport{

			Dial: dialer.Dial,

			DialContext: (&net.Dialer{
				Timeout: 10 * time.Second,
			}).DialContext,
		},
	}

	resp, err := client.Do(req)

	if err != nil {

		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	defer resp.Body.Close()

	for k, v := range resp.Header {

		for _, vv := range v {
			w.Header().Add(k, vv)
		}
	}

	w.WriteHeader(resp.StatusCode)

	io.Copy(w, resp.Body)
}

func transfer(destination io.WriteCloser, source io.ReadCloser) {

	defer destination.Close()
	defer source.Close()

	io.Copy(destination, source)
}

func showHelp() {

	fmt.Println()

	color.Cyan("════════════════ KYPROX HELP ════════════════")

	fmt.Println("\n1. Fetch proxies from:")
	fmt.Println("   https://libhub.stackverify.site/proxy")

	fmt.Println("\n2. Android WiFi Proxy Setup")

	fmt.Println("   Settings → WiFi")
	fmt.Println("   Long press network")
	fmt.Println("   Modify Network → Advanced")

	fmt.Println("\n   Proxy : Manual")
	fmt.Println("   Host  : 127.0.0.1")
	fmt.Println("   Port  : 8080")

	fmt.Println("\n3. Mobile Data Proxy (APN)")

	fmt.Println("   Settings → Mobile Network")
	fmt.Println("   Access Point Names")

	fmt.Println("\n   Proxy : 127.0.0.1")
	fmt.Println("   Port  : 8080")

	fmt.Println("\n4. Enable AdBlock from menu option 3")

	fmt.Println("\n═════════════════════════════════════════════")
}
