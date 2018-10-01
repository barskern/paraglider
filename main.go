package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	// Get port from environment
	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "8080"
	}

	log.Printf("Add handler for `/api`")
	http.HandleFunc("/api", MetaHandler)

	log.Printf("Setting up server to listen at port %s", port)

	// Run the server
	// This will block the current thread
	err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil)

	// We will only get to this statement if the server unexpectedly crashes
	log.Fatalf("Server error: %s", err)
}
