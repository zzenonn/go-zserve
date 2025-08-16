# go-zserve Documentation and Usage Guide

`go-zserve` is a Go-based utility that allows you to serve a file over HTTP on a specified port. It temporarily opens the port using `firewalld` and ensures that the port is closed when the program exits. 

**Note: This tool currently only works on Linux machines that use `firewall-cmd` (firewalld). It will not work on other operating systems or Linux distributions that use different firewall management tools.**

## Features

- Serves a specified file over HTTP.
- Temporarily opens a port using `firewalld` in a specified zone.
- Automatically closes the port when the program exits.
- Displays the file URL with the local machine's IP address.

## Requirements

- **Linux with firewalld**: This tool only works on Linux systems that use `firewall-cmd` (firewalld) for firewall management. It will not work on:
  - Windows or macOS
  - Linux distributions using `ufw`, `iptables`, or other firewall tools
- **Go**: You need to have Go installed to build and run this project.
- **firewalld**: The tool uses `firewalld` to open and close ports, so `firewalld` must be installed and running.
- **Root privileges**: The program must be run as root (with `sudo`) to modify firewall settings.
- **Task** (optional): For using the build tasks, install [Task](https://taskfile.dev/) task runner.

## Building from Source

### Using Task (Recommended)

If you have [Task](https://taskfile.dev/) installed, you can use the provided Taskfile.yml:

```bash
# Build the zserve binary (outputs to build/zserve)
go-task build

# Build for multiple platforms
go-task build-all

# Clean build artifacts
go-task clean

# Install to $GOPATH/bin
go-task install

# Run the application (requires sudo)
go-task run -- --port 8080 --zone public /path/to/file

# Format code, run tests, etc.
go-task fmt
go-task test
go-task vet
```

### Manual Build

To build manually using Go:

```bash
# Create build directory
mkdir -p build

# Build the binary
go build -o build/zserve main.go
```

### Cross-platform Builds

To build for different platforms:

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o build/zserve-linux-amd64 main.go

# Linux ARM64
GOOS=linux GOARCH=arm64 go build -o build/zserve-linux-arm64 main.go

# macOS AMD64
GOOS=darwin GOARCH=amd64 go build -o build/zserve-darwin-amd64 main.go

# macOS ARM64 (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o build/zserve-darwin-arm64 main.go
```

## Installation

### From Source

After building, you can install the binary:

```bash
# Using Task
go-task install

# Or manually
cp build/zserve $GOPATH/bin/zserve
```

### From GitHub

To install directly from GitHub:

```bash
go install github.com/zzenonn/go-zserve@latest
```

This will install the `go-zserve` binary into your `$GOPATH/bin` directory.

## Usage

### Basic Command

The basic command to run `go-zserve` is:

```bash
sudo zserve [flags] <file-path>
```

### Flags

- `-p, --port`: The port on which the file will be served (default: `8080`).
- `-z, --zone`: The `firewalld` zone to use (default: `public`).
- `-h, --help`: Show help message.

### Zone Parameter

The zone parameter specifies which firewalld zone to use for opening the port. Common zones include:

- **public** - Default zone, allows limited incoming connections
- **trusted** - Allows all network connections  
- **internal** - For internal networks with more trust
- **home** - For home networks

Use `firewall-cmd --get-zones` to see all available zones on your system.

### Examples

#### Basic Usage
```bash
# Serve a text file on default port 8080 in public zone
sudo zserve /path/to/file.txt
```

#### Custom Port and Zone
```bash
# Serve a PDF on port 9090 in trusted zone
sudo zserve --port 9090 --zone trusted /path/to/document.pdf

# Using short flags
sudo zserve -p 9090 -z trusted /path/to/document.pdf
```

#### Serving Large Files
```bash
# Serve a Linux ISO file for network installation
sudo zserve --port 8080 --zone public ~/Downloads/ubuntu-22.04.3-desktop-amd64.iso
```

#### Home Network Usage
```bash
# Serve an image file on home network
sudo zserve --port 8080 --zone home ~/Pictures/screenshot.png
```

### Help

To see all available options and examples:

```bash
zserve --help
# or
zserve -h
```

## How It Works

1. **Root Check**: The program checks if it is being run as root. If not, it exits with an error.
2. **File Check**: The program verifies if the provided file exists.
3. **Firewall Port Opening**: The program opens the specified port using `firewalld` in the specified zone.
4. **HTTP Server**: The program starts an HTTP server and serves the file at the specified port.
5. **Local IP Display**: The program retrieves the local IP address and prints the file URL.
6. **Cleanup**: When the program receives a termination signal (e.g., `Ctrl+C`), it closes the port in the firewall.

## Cleanup on Exit

The program listens for termination signals (e.g., `SIGINT`, `SIGTERM`) and ensures that the firewall port is closed before the program exits. This prevents the port from remaining open after the program has stopped.

## Troubleshooting

- **Permission Denied**: Ensure you are running the program with `sudo` or as root.
- **firewalld not running**: Ensure that `firewalld` is installed and running on your system.
- **Port already in use**: If the port is already in use, either stop the process using that port or specify a different port.
- **Unsupported OS/Firewall**: This tool only works on Linux systems with firewalld. It will not work on Windows, macOS, or Linux distributions using other firewall tools like `ufw` or `iptables`.
