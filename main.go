package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	// Use a flag for the port, but no flag for the file path
	port := flag.String("port", "8080", "The port to serve the file on")
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

	// Define the handler to serve the file based on its filename
	http.HandleFunc("/"+fileName, func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, absPath)
	})

	// Start the server and bind to all interfaces (0.0.0.0)
	go func() {
		log.Printf("Serving file %s on port %s\n", absPath, *port)
		if err := http.ListenAndServe("0.0.0.0:"+*port, nil); err != nil {
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

	// Keep the program running
	select {}
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
