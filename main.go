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
			fmt.Println("Usage: paragliding [-q][-v][-h]\n\n-q Quiet mode (only warn and error)\n-v Verbose mode (all logs)")
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

	// Create a new server which encompasses all state
	server := NewServer()

	// Route all requests to `paragliding/api/` to the api-server
	//
	// Remove prefix `/paragliding/api/` (api server shouldn't care where its
	// mounted)
	http.Handle("/paragliding/api/", http.StripPrefix("/paragliding/api", &server))

	// This function will block the current thread
	err := http.ListenAndServe(":"+port, nil)

	// We will only get to this statement if the server unexpectedly crashes
	log.WithFields(log.Fields{
		"cause": err,
	}).Fatal("server error occurred")
}
