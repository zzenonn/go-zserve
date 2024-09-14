package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
)

func main() {
	// Check if running as root (sudo)
	if !isRoot() {
		log.Fatal("This program must be run with sudo or as root")
	}

	// Use flags for the port and zone
	port := flag.String("port", "8080", "The port to serve the file on")
	zone := flag.String("zone", "public", "The firewalld zone to use")
	flag.Parse()

	// Ensure the file path is provided as an argument
	if len(flag.Args()) == 0 {
		log.Fatal("You must provide a file path as an argument")
	}

	// Get the file path from the arguments
	filePath := flag.Arg(0)

	// Check if the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Fatalf("The file %s does not exist", filePath)
	}

	// Get the absolute path of the file
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		log.Fatalf("Error getting absolute path: %v", err)
	}

	// Get the file name from the path
	fileName := filepath.Base(absPath)

	// Open the port temporarily using firewalld with the specified zone
	err = openFirewallPort(*port, *zone)
	if err != nil {
		log.Fatalf("Error opening firewall port: %v", err)
	}

	// Define the handler to serve the file based on its filename
	http.HandleFunc("/"+fileName, func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, absPath)
	})

	// Start the server and bind to all interfaces (0.0.0.0)
	go func() {
		log.Printf("Serving file %s on port %s\n", absPath, *port)
		if err := http.ListenAndServe(":"+*port, nil); err != nil {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	// Get the machine's local IP address
	ip, err := getLocalIP()
	if err != nil {
		log.Fatalf("Error getting local IP address: %v", err)
	}

	// Display the actual IP address and the file URL
	fmt.Printf("File is available at: http://%s:%s/%s\n", ip, *port, fileName)

	// Handle program termination signals to clean up
	cleanupOnExit(*port, *zone)

	// Keep the program running
	select {}
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

