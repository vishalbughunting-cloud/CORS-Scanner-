# CORS-Scanner-
CORS SCANNER ADVANCE 


 CORS Testing Tool 🔍

A powerful Go-based security tool for testing CORS misconfigurations and vulnerabilities in web applications.

![Go Version](https://img.shields.io/badge/Go-1.19+-blue)
![License](https://img.shields.io/badge/License-MIT-green)
![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20Linux%20%7C%20macOS-lightgrey)

## Features ✨

- **Single & Bulk URL Testing** - Test individual URLs or scan multiple targets from a file
- **CORS Vulnerability Detection** - Identify common CORS misconfigurations
- **Real-time Progress Bar** - Visual progress with ETA and statistics
- **Multiple Output Formats** - Detailed text reports with headers and vulnerabilities
- **High Performance** - Concurrent scanning with configurable threads
- **Cross-Platform** - Works on Windows, Linux, and macOS

## Vulnerabilities Detected 🚨

- 🔴 **Reflected Origin** - ACAO header mirrors the Origin header
- 🔴 **Wildcard with Credentials** - `*` with `Allow-Credentials: true`
- 🔴 **Simple Wildcard** - `*` without proper restrictions
- 🔴 **Null Origin** - `null` origin with credentials enabled
- 🔴 **Prefix Matching** - Weak origin validation patterns
- 🔴 **Multiple Origins** - Comma-separated origin values



# Clone the repository
git clone https://github.com/vishalbughunting-cloud/cors-tool.git
cd cors-tool
go build -o cors-tool.exe main.go
Now Ready To Scan 


Single URL with Vulnerability Detection:
cors-tool.exe -url https://example.com -verbose


Bulk Scan with Progress Bar:

cors-tool.exe -file urls.txt -output scan_results.txt



