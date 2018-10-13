package main

import (
	"encoding/json"
	"github.com/marni/goigc"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Methods the server will handle
var allowedMethods = [...]string{http.MethodPost, http.MethodGet}

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
	logger := log.WithFields(log.Fields{
		"method": r.Method,
		"path":   r.URL.Path,
		"addr":   r.RemoteAddr,
	})

	logger.Info("processing request")
	switch r.Method {
	case http.MethodGet:
		switch r.URL.Path {
		case "/":
			logger.Info("forwared request to metadata handler")
			server.metaHandler(logger, w, r)
		default:
			logger.Info("path not found, responding with 404 (not found)")
			http.NotFound(w, r)
		}
	case http.MethodPost:
		switch r.URL.Path {
		case "/igc":
			logger.Info("forwared request to track registration handler")
			server.trackRegHandler(logger, w, r)
		default:
			logger.Info("path not found, responding with 404 (not found)")
			http.NotFound(w, r)
		}
	default:
		logger.Info("invalid method, responding with 405 (status method not allowed)")
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
func (server *IgcServer) metaHandler(logger *log.Entry, w http.ResponseWriter, _ *http.Request) {
	metadata := map[string]interface{}{
		"uptime":  FormatAsISO8601(time.Since(server.StartupTime)),
		"info":    "Service for IGC tracks.",
		"version": "v1",
	}

	logger.WithFields(log.Fields(metadata)).Info("responding with metadata")

	// Encode metadata as a JSON object
	json.NewEncoder(w).Encode(metadata)
}

// trackRegHandler takes a request in the following structure
//
// ```json
// {
//   "url": <some-url>
// }
// ```
//
// If a valid url to a `.igc` file is provided, the response will be in the
// following structure
//
// ```json
// {
//   "id": <TrackID>
// }
// ```
// FIXME errors are handled gracefully but verbosly, is there a better way?
func (server *IgcServer) trackRegHandler(logger *log.Entry, w http.ResponseWriter, r *http.Request) {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var req TrackRegRequest
	if err := dec.Decode(&req); err != nil {
		logger.WithField("error", err).Info("unable to decode request body")
		http.Error(w, "invalid json object", http.StatusBadRequest)
		return
	}
	reqURL, err := url.Parse(req.URLstr)
	if err != nil {
		logger.WithField("error", err).Info("unable to parse url")
		http.Error(w, "invalid url", http.StatusBadRequest)
		return
	}
	resp, err := http.Get(reqURL.String())
	if err != nil {
		logger.WithField("error", err).Info("unable to fetch data from provided url")
		http.Error(w, "unable to fetch data from provided url", http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.WithField("error", err).Error("unable to read all data from response")
		http.Error(w, "unable to read data from provided url", http.StatusInternalServerError)
		return
	}
	track, err := igc.Parse(string(content))
	if err != nil {
		logger.WithField("error", err).Info("unable to parse igc content as track")
		// igc parse error is a returned as bad request error code because the
		// file that the user linked to contains invalid igc content
		http.Error(w, "unable to parse igc content", http.StatusBadRequest)
		return
	}

	// Create and add new trackmeta object
	trackMeta := TrackMetaFrom(track)
	id := server.Data.Append(trackMeta)

	result := map[string]interface{}{
		"id": id,
	}

	logger.WithFields(log.Fields{
		"trackmeta": trackMeta,
		"id":        id,
	}).Info("responding with id of inserted track metadata")

	json.NewEncoder(w).Encode(result)
}

// TrackRegRequest is the format of a track registration request
type TrackRegRequest struct {
	URLstr string `json:"url"`
}
