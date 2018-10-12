package main

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

// startupTime contains the time the application started running
var startupTime time.Time

// Initializes `startupTime` so that we can later calculate uptime
func init() {
	startupTime = time.Now()

	log.WithFields(log.Fields{
		"value": startupTime,
	}).Debug("initialized startupTime")
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

	log.WithFields(log.Fields{
		"value": time.Duration(uptime),
	}).Debug("initialized uptime")

	metadata := metadata{uptime, "Service for IGC tracks.", "v1"}

	// Encode metadata as a JSON object
	json.NewEncoder(w).Encode(metadata)
}
