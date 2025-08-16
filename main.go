package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"
)

var (
	port string
	zone string
)

var rootCmd = &cobra.Command{
	Use:   "zserve [flags] <file-path>",
	Short: "Serve a file over HTTP with automatic firewall management",
	Long: `zserve temporarily serves a file over HTTP on a specified port.
It automatically opens the port using firewalld and closes it when
the program exits. Must be run with sudo or as root.

The zone parameter specifies which firewalld zone to use for opening
the port. Common zones include:
  public   - Default zone, allows limited incoming connections
  trusted  - Allows all network connections  
  internal - For internal networks with more trust
  home     - For home networks

Use 'firewall-cmd --get-zones' to see all available zones.`,
	Example: `  # Serve a text file on default port 8080
  sudo zserve /path/to/file.txt

  # Serve a PDF on port 9090 in trusted zone
  sudo zserve --port 9090 --zone trusted /path/to/document.pdf

  # Serve a Linux ISO file for network installation
  sudo zserve --port 8080 --zone public ~/Downloads/ubuntu-22.04.3-desktop-amd64.iso

  # Serve an image file on home network
  sudo zserve --port 8080 --zone home ~/Pictures/screenshot.png`,
	Args: cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Check if running as root (sudo)
		if !isRoot() {
			return fmt.Errorf("this program must be run with sudo or as root\n\nTry: sudo zserve [flags] <file-path>")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]

		// Check if the file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return fmt.Errorf("the file %s does not exist", filePath)
		}

		// Get the absolute path of the file
		absPath, err := filepath.Abs(filePath)
		if err != nil {
			return fmt.Errorf("error getting absolute path: %v", err)
		}

		// Get the file name from the path
		fileName := filepath.Base(absPath)

		// Open the port temporarily using firewalld with the specified zone
		err = openFirewallPort(port, zone)
		if err != nil {
			return fmt.Errorf("error opening firewall port: %v", err)
		}

		// Define the handler to serve the file based on its filename
		http.HandleFunc("/"+fileName, func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, absPath)
		})

		// Start the server and bind to all interfaces (0.0.0.0)
		go func() {
			log.Printf("Serving file %s on port %s\n", absPath, port)
			if err := http.ListenAndServe(":"+port, nil); err != nil {
				log.Fatalf("Error starting server: %v", err)
			}
		}()

		// Get the machine's local IP address
		ip, err := getLocalIP()
		if err != nil {
			return fmt.Errorf("error getting local IP address: %v", err)
		}

		// Display the actual IP address and the file URL
		fmt.Printf("File is available at: http://%s:%s/%s\n", ip, port, fileName)
		fmt.Printf("Press Ctrl+C to stop the server and close the firewall port.\n")

		// Handle program termination signals to clean up
		cleanupOnExit(port, zone)

		// Keep the program running
		select {}
	},
}

func init() {
	rootCmd.Flags().StringVarP(&port, "port", "p", "8080", "The port to serve the file on (default: 8080)")
	rootCmd.Flags().StringVarP(&zone, "zone", "z", "public", "The firewalld zone to use (default: public)")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// isRoot checks if the program is being run as root (with sudo)
func isRoot() bool {
	return os.Geteuid() == 0
}

// openFirewallPort opens the specified port temporarily using firewalld in the given zone
func openFirewallPort(port string, zone string) error {
	// Run the firewall-cmd command to open the port temporarily in the specified zone
	cmd := exec.Command("firewall-cmd", "--zone="+zone, "--add-port="+port+"/tcp")
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to open firewall port in zone %s: %w", zone, err)
	}

	log.Printf("Firewall port %s opened temporarily in zone %s", port, zone)
	return nil
}

// closeFirewallPort removes the specified port using firewalld in the given zone
func closeFirewallPort(port string, zone string) error {
	// Run the firewall-cmd command to remove the port in the specified zone
	cmd := exec.Command("firewall-cmd", "--zone="+zone, "--remove-port="+port+"/tcp")
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to remove firewall port in zone %s: %w", zone, err)
	}

	log.Printf("Firewall port %s removed in zone %s", port, zone)
	return nil
}

// cleanupOnExit handles cleanup when the program exits by catching signals and removing the port
func cleanupOnExit(port string, zone string) {
	// Set up channel to listen for interrupt or termination signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("Received exit signal, removing firewall port.")
		if err := closeFirewallPort(port, zone); err != nil {
			log.Fatalf("Error removing firewall port: %v", err)
		}
		os.Exit(0)
	}()
}

// getLocalIP retrieves the actual local IP address of the machine
func getLocalIP() (string, error) {
	// Create a UDP connection to a public IP (doesn't need to connect)
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	// Get the local IP address from the UDP connection
	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String(), nil
}
