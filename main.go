package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func hello_world(w http.ResponseWriter, r *http.Request) {
	// Make sure request isn't nil before accessing fields
	if r != nil {
		log.Printf("Recived a %s-request from %s for %s", r.Method, r.URL, r.RemoteAddr)
	} else {
		log.Printf("Recived a request which is nil")
	}
	fmt.Fprintf(w, "Hello world")
}

func main() {
	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "8080"
	}

	log.Print("Adding `hello_world` handler to root path")
	http.HandleFunc("/", hello_world)

	log.Printf("Setting up server to listen at port %s", port)

	// Run the server
	// This will block the current thread
	err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil)

	// We will only get to this statement if the server unexpectedly crashes
	log.Fatalf("Server error: %s", err)
}
