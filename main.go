package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	// Define the flag for file path input
	filePath := flag.String("file", "", "The path to the file to serve")
	port := flag.String("port", "8080", "The port to serve the file on")
	flag.Parse()

	// Ensure a file path is provided
	if *filePath == "" {
		log.Fatal("You must provide a file path using the -file flag")
	}

	// Check if the file exists
	if _, err := os.Stat(*filePath); os.IsNotExist(err) {
		log.Fatalf("The file %s does not exist", *filePath)
	}

	// Get the absolute path of the file
	absPath, err := filepath.Abs(*filePath)
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

	// Display the URL for accessing the file
	fmt.Printf("File is available at: http://localhost:%s/%s\n", *port, fileName)

	// Keep the program running
	select {}
}

