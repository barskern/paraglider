package main

import (
	log "github.com/sirupsen/logrus"
	"net/http"
	"sync"
	"time"
)

const allowedMethods = "GET"

// A unique id for a track
// TODO give a meaningful type, perhaps a generated uuid
type trackID int

// TrackMeta contains a subset of metainformation about a igc-track
//
// ```json
// {
//   "H_date": <date from File Header, H-record>,
//   "pilot": <pilot>,
//   "glider": <glider>,
//   "glider_id": <glider_id>,
//   "track_length": <calculated total track length>
// }
// ```
type TrackMeta struct {
	Date        time.Time `json:"H_date"`
	Pilot       string    `json:"pilot"`
	Gilder      string    `json:"glider"`
	GliderID    string    `json:"glider_id"`
	TrackLength float64   `json:"track_length"`
}

// TODO make function which converts a igc.Track into a TrackMeta struct.
// See https://godoc.org/github.com/marni/goigc#Track

// TrackMetas contains a map to many TrackMeta objects which are protected
// by a RWMutex and indexed by a unique id
type TrackMetas struct {
	mu   sync.RWMutex
	data map[trackID]TrackMeta
}

// TODO implement a read and write function which returns the map and the mutex
// after locking

// NewTrackMetas creates a new mutex and mapping from ID to TrackMeta
func NewTrackMetas() TrackMetas {
	return TrackMetas{sync.RWMutex{}, make(map[trackID]TrackMeta)}
}

// IgcServer distributes request to a pool of worker gorutines
type IgcServer struct {
	data TrackMetas
}

// NewIgcServer creates a new server which handles requests to the igc api
func NewIgcServer() IgcServer {
	return IgcServer{NewTrackMetas()}
}

func (server *IgcServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	reqLog := log.WithFields(log.Fields{
		"method": r.Method,
		"path":   r.URL.Path,
		"addr":   r.RemoteAddr,
	})

	reqLog.Info("processing request")
	switch r.Method {
	case "GET":
		switch r.URL.Path {
		case "", "/":
			reqLog.Info("forwared request to metadata handler")
			MetaHandler(w, r)
		default:
			reqLog.Info("path not found, responding with 404 (not found)")
			http.NotFound(w, r)
		}
	default:
		reqLog.Info("invalid method, responding with 405 (status method not allowed)")
		// A 405 response MUST generate an 'Allow' header which specifies the
		// methods that are valid (RFC7231 6.5.5)
		w.Header().Add("Allow", allowedMethods)
		http.Error(w, "status method is not allowed", http.StatusMethodNotAllowed)
	}
}
