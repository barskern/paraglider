package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
)

func main() {
	for _, v := range os.Args {
		switch v {
		case "-v":
			log.SetLevel(log.DebugLevel)
		case "-q":
			log.SetLevel(log.WarnLevel)
		case "-h":
			fmt.Println("Usage: igcinfo [-q][-v][-h]\n\n-q Quiet mode (only warn and error)\n-v Verbose mode (all logs)")
			os.Exit(0)
		}
	}

	// Get port from environment
	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "8080"
	}

	log.WithFields(log.Fields{
		"port":     port,
		"logLevel": log.GetLevel(),
	}).Info("initializing server")

	// Create a new server which encompasses all server state
	igcServer := NewIgcServer()

	// Route all requests to `igcinfo/api/` to the igc-server
	//
	// Remove the `/igcinfo/api/` so that the server can handle requests directly
	// without caring about the api-point its mounted on
	http.Handle("/igcinfo/api/", http.StripPrefix("/igcinfo/api", &igcServer))

	// Run the server
	// This function will block the current thread
	err := http.ListenAndServe(":"+port, nil)

	// We will only get to this statement if the server unexpectedly crashes
	log.WithFields(log.Fields{
		"cause": err,
	}).Fatal("server error occurred")
}
