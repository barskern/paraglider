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

	log.Info("Add handler for `/api`")
	http.HandleFunc("/api", MetaHandler)

	log.WithFields(log.Fields{
		"value": port,
	}).Info("Setting up server to listen at port")

	// Run the server
	// This funciton will block the current thread
	err := http.ListenAndServe(":"+port, nil)

	// We will only get to this statement if the server unexpectedly crashes
	log.WithFields(log.Fields{
		"cause": err,
	}).Fatal("server error occured")
}
