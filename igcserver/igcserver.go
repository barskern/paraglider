package igcserver

import (
	"encoding/json"
	"github.com/barskern/paragliding/isodur"
	"github.com/gorilla/mux"
	"github.com/marni/goigc"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// Server distributes request to a pool of worker gorutines
type Server struct {
	startupTime time.Time
	data        TrackMetas
	httpClient  *http.Client
	// TODO change from mux to pure regexes because of the simple routing
	router *mux.Router
}

// NewServer creates a new server which handles requests to the igc api
func NewServer(httpClient *http.Client, trackMetas TrackMetas) (srv Server) {
	srv = Server{
		time.Now(),
		trackMetas,
		httpClient,
		mux.NewRouter(),
	}
	srv.router.HandleFunc("/", srv.metaHandler).Methods(http.MethodGet)
	srv.router.HandleFunc("/track", srv.trackRegHandler).Methods(http.MethodPost)
	srv.router.HandleFunc("/track", srv.trackGetAllHandler).Methods(http.MethodGet)

	srv.router.HandleFunc(
		"/track/{id}",
		srv.trackGetHandler,
	).Methods(http.MethodGet)

	srv.router.HandleFunc(
		"/track/{id}/{field}",
		srv.trackGetFieldHandler,
	).Methods(http.MethodGet)

	srv.router.Use(loggingMiddleware)

	srv.router.MethodNotAllowedHandler =
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := newReqLogger(r)
			logger.Info("received request with disallowed method")

			// A 405 MUST generate "Allow" header in the header (rfc 7231 6.5.5)
			w.Header().Add("Allow", "GET POST")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		})

	srv.router.NotFoundHandler =
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := newReqLogger(r)
			logger.Info("received request which didn't match any paths")

			http.Error(w, "content not found", http.StatusNotFound)
		})

	return
}

func (server *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

// metaHandler returns the metadata about the api endpoint
func (server *Server) metaHandler(w http.ResponseWriter, r *http.Request) {
	logger := newReqLogger(r)

	metadata := map[string]interface{}{
		"uptime":  isodur.FormatAsISO8601(time.Since(server.startupTime)),
		"info":    "Service for Paragliding tracks.",
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
func (server *Server) trackRegHandler(w http.ResponseWriter, r *http.Request) {
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
	resp, err := server.httpClient.Get(reqURL.String())
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
	trackMeta := TrackMetaFrom(*reqURL, track)
	err = server.data.Append(trackMeta)
	if err == ErrTrackAlreadyExists {
		logger.WithFields(log.Fields{
			"trackmeta": trackMeta,
		}).Info("request attempted to add duplicate track metadata")
		http.Error(w, "track with same url already exists", http.StatusForbidden)
	} else if err != nil {
		logger.WithFields(log.Fields{
			"trackmeta": trackMeta,
			"error":     err,
		}).Info("unable to add track metadata")
		http.Error(w, "internal server error occured", http.StatusInternalServerError)
		return
	}

	result := map[string]interface{}{
		"id": trackMeta.ID,
	}

	logger.WithFields(log.Fields{
		"trackmeta": trackMeta,
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
func (server *Server) trackGetAllHandler(w http.ResponseWriter, r *http.Request) {
	logger := newReqLogger(r)

	ids, err := server.data.GetAllIDs()
	if err != nil {
		logger.WithField("error", err).Error("unable to respond to request of all IDs")
		http.Error(w, "internal server error occured", http.StatusInternalServerError)
		return
	}
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
//   "track_length": <calculated total track length>,
//   "track_src_url": <the original URL used to upload the track, ie. the URL used with POST>
// }
//```
func (server *Server) trackGetHandler(w http.ResponseWriter, r *http.Request) {
	logger := newReqLogger(r)

	vars := mux.Vars(r)
	// Should never fail because of mux
	idStr, _ := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.WithField("id", idStr).Info("id must be a valid number")
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	idlog := logger.WithField("id", id)
	meta, err := server.data.Get(TrackID(id))
	if err == ErrTrackNotFound {
		idlog.Info("unable to find metadata of id")
		http.Error(w, "content not found", http.StatusNotFound)
		return
	} else if err != nil {
		idlog.WithField("error", err).Info("error when getting metadata of id")
		http.Error(w, "internal server error occured", http.StatusInternalServerError)
		return
	}
	logger.WithFields(log.Fields{
		"trackmeta": meta,
	}).Info("responding with track meta for given id")
	json.NewEncoder(w).Encode(meta)
}

// trackGetFieldHandler should return the field specified in the url
func (server *Server) trackGetFieldHandler(w http.ResponseWriter, r *http.Request) {
	logger := newReqLogger(r)

	vars := mux.Vars(r)

	// Should never fail because of mux
	idStr, _ := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.WithField("id", idStr).Info("id must be a valid number")
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	field, _ := vars["field"]
	idlog := logger.WithField("id", id)

	meta, err := server.data.Get(TrackID(id))
	if err == ErrTrackNotFound {
		idlog.Info("unable to find metadata of id")
		http.Error(w, "content not found", http.StatusNotFound)
		return
	} else if err != nil {
		idlog.WithField("error", err).Info("error when getting metadata of id")
		http.Error(w, "internal server error occured", http.StatusInternalServerError)
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
	case "track_src_url":
		flog.Info("responding with track src url")
		w.Write([]byte(meta.TrackSrcURL))
	default:
		flog.Info("unable to find field of metadata")
		http.Error(w, "invalid field", http.StatusBadRequest)
	}
}
