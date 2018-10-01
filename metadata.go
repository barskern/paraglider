package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

var DEFAULT_METADATA metadata
var STARTUP_TIME time.Time

// Initializes STARTUP_TIME so that `Uptime` can be calculated
func init() {
	STARTUP_TIME = time.Now()
	DEFAULT_METADATA = metadata{0, "Service for IGC tracks.", "v1"}
}

// Contains the metadata which is sent back to the user when requesting `/api`
type metadata struct {
	Uptime  JsonDuration `json:"uptime"`
	Info    string       `json:"info"`
	Version string       `json:"version"`
}

// Return metadata about the api endpoint
//
// Should be in the following format:
// {
//   "uptime": <uptime>
//   "info": "Service for IGC tracks."
//   "version": "v1"
// }
func MetaHandler(w http.ResponseWriter, r *http.Request) {
	// Make local metadata
	metadata := DEFAULT_METADATA
	// Update uptime for local metadata
	metadata.Uptime = JsonDuration(time.Since(STARTUP_TIME))
	log.Printf("sending metadata")
	// Encode metadata as response
	json.NewEncoder(w).Encode(metadata)
}

type JsonDuration time.Duration

func (t JsonDuration) MarshalJSON() ([]byte, error) {
	d := time.Duration(t)
	dur_str := fmt.Sprintf("\"P%dY%dM%dDT%dH%dM%dS\"", 0, 0, 0, int(d.Hours()), int(d.Minutes()), int(d.Seconds()))
	return []byte(dur_str), nil
}
