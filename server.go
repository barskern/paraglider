package main

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
	"time"
)

// Methods the server will handle
var allowedMethods = [...]string{http.MethodGet}

// IgcServer distributes request to a pool of worker gorutines
type IgcServer struct {
	StartupTime time.Time
	Data        TrackMetas
}

// NewIgcServer creates a new server which handles requests to the igc api
func NewIgcServer() IgcServer {
	return IgcServer{time.Now(), NewTrackMetas()}
}

func (server *IgcServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	reqLog := log.WithFields(log.Fields{
		"method": r.Method,
		"path":   r.URL.Path,
		"addr":   r.RemoteAddr,
	})

	reqLog.Info("processing request")
	switch r.Method {
	case http.MethodGet:
		switch r.URL.Path {
		case "/":
			reqLog.Info("forwared request to metadata handler")
			server.metaHandler(w, r)
		default:
			reqLog.Info("path not found, responding with 404 (not found)")
			http.NotFound(w, r)
		}
	default:
		reqLog.Info("invalid method, responding with 405 (status method not allowed)")
		// A 405 response MUST generate an 'Allow' header which specifies the
		// methods that are valid (RFC7231 6.5.5)
		w.Header().Add("Allow", strings.Join(allowedMethods[:], " "))
		http.Error(w, "status method is not allowed", http.StatusMethodNotAllowed)
	}
}

// metaHandler returns the metadata about the api endpoint in the following
// structure
//
// ```json
// {
//   "uptime": <uptime>
//   "info": "Service for IGC tracks."
//   "version": "v1"
// }
// ```
func (server *IgcServer) metaHandler(w http.ResponseWriter, _ *http.Request) {
	metadata := map[string]interface{}{
		"uptime":  FormatAsISO8601(time.Since(server.StartupTime)),
		"info":    "Service for IGC tracks.",
		"version": "v1",
	}

	log.WithFields(log.Fields(metadata)).Info("responding with metadata")

	// Encode metadata as a JSON object
	json.NewEncoder(w).Encode(metadata)
}
