package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// startupTime contains the time the application started running
var startupTime time.Time

// Initializes `startupTime` so that we can later calculate uptime
func init() {
	startupTime = time.Now()
}

// metadata encodes the structure of metadata
type metadata struct {
	Uptime  ISO8601Duration `json:"uptime"`
	Info    string          `json:"info"`
	Version string          `json:"version"`
}

// MetaHandler returns the metadata about the api endpoint in the following
// structure
//
// ```json
// {
//   "uptime": <uptime>
//   "info": "Service for IGC tracks."
//   "version": "v1"
// }
// ```
func MetaHandler(w http.ResponseWriter, _ *http.Request) {
	// Make local metadata with current uptime
	uptime := ISO8601Duration(time.Since(startupTime))
	metadata := metadata{uptime, "Service for IGC tracks.", "v1"}

	log.Printf("responing to request for metadata")

	// Encode metadata as a JSON object
	json.NewEncoder(w).Encode(metadata)
}
