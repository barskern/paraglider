package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/marni/goigc"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// IgcServer distributes request to a pool of worker gorutines
type IgcServer struct {
	startupTime time.Time
	data        TrackMetas
	router      *mux.Router
}

// NewIgcServer creates a new server which handles requests to the igc api
func NewIgcServer() (srv IgcServer) {
	srv = IgcServer{
		time.Now(),
		NewTrackMetas(),
		mux.NewRouter(),
	}
	srv.router.HandleFunc("/", srv.metaHandler).Methods(http.MethodGet)
	srv.router.HandleFunc("/igc", srv.trackRegHandler).Methods(http.MethodPost)
	srv.router.HandleFunc("/igc", srv.trackGetAllHandler).Methods(http.MethodGet)

	srv.router.HandleFunc(
		"/igc/{id:[A-Za-z0-9+/]{8}}",
		srv.trackGetHandler,
	).Methods(http.MethodGet)

	srv.router.HandleFunc(
		"/igc/{id:[A-Za-z0-9+/]{8}}/{field:[a-zA-Z0-9_-]+}",
		srv.trackGetFieldHandler,
	).Methods(http.MethodGet)

	srv.router.Use(loggingMiddleware)

	srv.router.MethodNotAllowedHandler =
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := newReqLogger(r)
			logger.Info("recieved request with disallowed method")

			// A 405 MUST generate "Allow" header in the header (rfc 7231 6.5.5)
			w.Header().Add("Allow", "GET POST")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		})

	srv.router.NotFoundHandler =
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := newReqLogger(r)
			logger.Info("recieved request which didn't match any paths")

			http.Error(w, "content not found", http.StatusNotFound)
		})

	return
}

func (server *IgcServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	server.router.ServeHTTP(w, r)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := newReqLogger(r)
		logger.Info("received request")
		next.ServeHTTP(w, r)
	})
}

func newReqLogger(r *http.Request) *log.Entry {
	return log.WithFields(log.Fields{
		"method": r.Method,
		"path":   r.URL.Path,
		"addr":   r.RemoteAddr,
	})
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
func (server *IgcServer) metaHandler(w http.ResponseWriter, r *http.Request) {
	logger := newReqLogger(r)

	metadata := map[string]interface{}{
		"uptime":  FormatAsISO8601(time.Since(server.startupTime)),
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
func (server *IgcServer) trackRegHandler(w http.ResponseWriter, r *http.Request) {
	logger := newReqLogger(r)

	logger.Info("processing request to register track")

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
	id := server.data.Append(trackMeta)

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

// trackGetAllHandler returns all ids of registered igc files in the structure
//
// ```json
// [ <id1>, <id2>, ... ]
// ```
func (server *IgcServer) trackGetAllHandler(w http.ResponseWriter, r *http.Request) {
	logger := newReqLogger(r)

	ids := server.data.GetAllIDs()
	logger.WithField("ids", ids).Info("responding to request with all ids")
	json.NewEncoder(w).Encode(ids)
}

// trackGetHandler should return the fields of a specific id in the structure
//
// ```json
// {
//   "H_date": <date from File Header, H-record>,
//   "pilot": <pilot>,
//   "glider": <glider>,
//   "glider_id": <glider_id>,
//   "track_length": <calculated total track length>
// }
//```
func (server *IgcServer) trackGetHandler(w http.ResponseWriter, r *http.Request) {
	logger := newReqLogger(r)

	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		logger.Info("unable to find id in vars of request")
		http.Error(w, "invalid or missing id", http.StatusBadRequest)
		return
	}
	meta, ok := server.data.Get(TrackID(id))
	if !ok {
		logger.WithField("id", id).Info("unable to find metadata of id")
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	logger.WithFields(log.Fields{
		"trackmeta": meta,
		"id":        id,
	}).Info("responding with track meta for given id")
	json.NewEncoder(w).Encode(meta)
}

// trackGetFieldHandler should return the field specified in the url
func (server *IgcServer) trackGetFieldHandler(w http.ResponseWriter, r *http.Request) {
	logger := newReqLogger(r)

	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		logger.Info("unable to find id in vars of request")
		http.Error(w, "invalid or missing id", http.StatusBadRequest)
		return
	}
	field, ok := vars["field"]
	if !ok {
		logger.Info("unable to find field in vars of request")
		http.Error(w, "invalid or missing field", http.StatusBadRequest)
		return
	}
	idlog := logger.WithField("id", id)
	meta, ok := server.data.Get(TrackID(id))
	if !ok {
		idlog.Info("unable to find metadata of id")
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	flog := idlog.WithField("field", field)
	switch field {
	case "H_date":
		flog.Info("responding with track date")
		w.Write([]byte(meta.Date.Format(time.RFC3339)))
	case "pilot":
		flog.Info("responding with track pilot")
		w.Write([]byte(meta.Pilot))
	case "glider":
		flog.Info("responding with track glider")
		w.Write([]byte(meta.Glider))
	case "glider_id":
		flog.Info("responding with track glider id")
		w.Write([]byte(meta.GliderID))
	case "track_length":
		flog.Info("responding with track length")
		w.Write([]byte(strconv.FormatFloat(meta.TrackLength, 'f', -1, 64)))
	default:
		flog.Info("unable to find field of metadata")
		http.Error(w, "invalid field", http.StatusBadRequest)
	}
}
