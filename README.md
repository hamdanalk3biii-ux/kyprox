# KYPROX - Termux Proxy Manager

**KYPROX** is a lightweight, easy-to-use proxy manager for Termux. It allows users to fetch SOCKS5 proxies, manually enter proxies, toggle ad-blocking, and route traffic through a local proxy without needing Go or any external files.

---

## Features

- Fetch free SOCKS5 proxies from [LibHub StackVerify](https://libhub.stackverify.site/proxy)
- Manual proxy entry (SOCKS5)
- Built-in ad-block (ads.json compiled into binary)
- Local HTTP proxy running on **127.0.0.1:8080**
- Easy-to-use menu interface with colors and tables
- Works on mobile Termux environment
- No dependencies required on the user side

---

## Installation (Termux)

Simply run the following one-line command in Termux to install KYPROX:

```bash
curl -L https://raw.githubusercontent.com/Frost-bit-star/kyprox/main/kyprox -o $PREFIX/bin/kyprox && chmod +x $PREFIX/bin/kyprox && echo 'alias kyprox="$PREFIX/bin/kyprox"' >> ~/.bashrc && echo 'alias kyprox="$PREFIX/bin/kyprox"' >> ~/.zshrc && source ~/.bashrc
```
## This will:

Download the compiled binary into Termux.
Make it executable.

Add an alias so you can run kyprox from anywhere.

## Usage
Open Termux.
- Run KYPROX:
Bash
```
kyprox
```
## You’ll see the menu:

1. Fetch proxy from API
2. Enter your own proxy
3. Toggle Ads Block
4. Help
5. Exit
   
## Important: Set your mobile/Wi-Fi APN proxy first so traffic is routed through KYPROX:
- Mobile / APN Settings:
- Host: 127.0.0.1
- Port: 8080
## Wi-Fi Settings:
- Go to your network → Advanced → Proxy → Manual
- Host: 127.0.0.1
- Port: 8080
## Select the option you want from the menu and KYPROX will handle the rest.
- Ad-block
- KYPROX comes with a built-in ads.json.
- Toggle ad-blocking from the menu option 3.
- When enabled, KYPROX will block requests to known ad domains.
## Notes
Currently, only SOCKS5 proxies are supported.
SOCKS4 is planned but not implemented yet.
Always set the proxy in your mobile/Wi-Fi before using KYPROX.
Compatible with Termux on Android.
## Support
For issues, bug reports, or suggestions, please open an issue on the GitHub repository:
