# go-zserve Documentation and Usage Guide

`go-zserve` is a Go-based utility that allows you to serve a file over HTTP on a specified port. It temporarily opens the port using `firewalld` and ensures that the port is closed when the program exits. I made this for my personal use with my system in mind.

## Features

- Serves a specified file over HTTP.
- Temporarily opens a port using `firewalld` in a specified zone.
- Automatically closes the port when the program exits.
- Displays the file URL with the local machine's IP address.

## Requirements

- **Go**: You need to have Go installed to build and run this project.
- **firewalld**: The tool uses `firewalld` to open and close ports, so `firewalld` must be installed and running.
- **Root privileges**: The program must be run as root (with `sudo`) to modify firewall settings.

## Installation

To install `go-zserve`, you can use the `go install` command:

```bash
go install github.com/zzenonn/go-zserve@latest
```

This will install the `go-zserve` binary into your `$GOPATH/bin` directory.

## Usage

### Basic Command

The basic command to run `go-zserve` is:

```bash
sudo go-zserve -port <port> -zone <zone> <file-path>
```

- **port**: The port on which the file will be served (default: `8080`).
- **zone**: The `firewalld` zone to use (default: `public`).
- **file-path**: The path to the file you want to serve.

### Example

To serve a file named `example.txt` on port `8080` in the `public` zone:

```bash
sudo go-zserve -port 8080 -zone public /path/to/example.txt
```

Once the server starts, the file will be accessible via a URL like:

```
http://<local-ip>:8080/example.txt
```

Where `<local-ip>` is the local IP address of your machine.

### Flags

- `-port`: The port number on which the file will be served. Default is `8080`.
- `-zone`: The `firewalld` zone in which the port will be opened. Default is `public`.

### Example with Custom Port and Zone

To serve a file on port `9090` in the `trusted` zone:

```bash
sudo go-zserve -port 9090 -zone trusted /path/to/example.txt
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
