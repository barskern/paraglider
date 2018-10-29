package igcserver

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"github.com/marni/goigc"
	log "github.com/sirupsen/logrus"
	"hash/fnv"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

var (
	// ErrTrackNotFound is returned if a request did not result in a TrackMeta
	ErrTrackNotFound = errors.New("track not found")

	// ErrTrackAlreadyExists is returned to request to add a track which
	// already exists
	ErrTrackAlreadyExists = errors.New("track already exists")
)

// TrackMetas is a interface for all storages containing TrackMeta
type TrackMetas interface {
	Get(id TrackID) (TrackMeta, error)
	Append(meta TrackMeta) error
	GetAllIDs() ([]TrackID, error)
}

// TrackID is a unique id for a track
type TrackID uint32

// NewTrackID creates a new unique track ID
func NewTrackID(v []byte) TrackID {
	hasher := fnv.New32()
	hasher.Write(v)
	return TrackID(hasher.Sum32())
}

// TrackMeta contains a subset of metainformation about a igc-track
type TrackMeta struct {
	ID          TrackID   `json:"-" bson:"id"`
	Timestamp   time.Time `json:"-" bson:"timestamp"`
	Date        time.Time `json:"H_date" bson:"H_date"`
	Pilot       string    `json:"pilot" bson:"pilot"`
	Glider      string    `json:"glider" bson:"glider"`
	GliderID    string    `json:"glider_id" bson:"glider_id"`
	TrackLength float64   `json:"track_length" bson:"track_length"`
	TrackSrcURL string    `json:"track_src_url" bson:"track_src_url"`
}

// calcTotalDistance returns the total distance between the points in order
func calcTotalDistance(points []igc.Point) (trackLength float64) {
	for i := 0; i+1 < len(points); i++ {
		trackLength += points[i].Distance(points[i+1])
	}
	return
}

// TrackMetaFrom converts a igc.Track into a TrackMeta struct
func TrackMetaFrom(url url.URL, track igc.Track) TrackMeta {
	return TrackMeta{
		NewTrackID([]byte(url.String())),
		time.Now(),
		track.Date,
		track.Pilot,
		track.GliderType,
		track.GliderID,
		calcTotalDistance(track.Points),
		url.String(),
	}
}

// --------- //
// TRACK API //
// --------- //

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
	// Check if track already exists before requesting an external service to
	// prevent unnecessary external calls
	id := NewTrackID([]byte(reqURL.String()))
	_, err = server.tracks.Get(id)
	if err == nil {
		logger.Info("request attempted to add duplicate track metadata")
		http.Error(w, "track with same url already exists", http.StatusForbidden)
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
		http.Error(w, "unable to parse igc content", http.StatusBadRequest)
		return
	}

	// Create and add new trackmeta object
	trackMeta := TrackMetaFrom(*reqURL, track)
	err = server.tracks.Append(trackMeta)
	if err == ErrTrackAlreadyExists {
		logger.WithFields(log.Fields{
			"trackmeta": trackMeta,
		}).Info("request attempted to add duplicate track metadata")
		http.Error(w, "track with same url already exists", http.StatusForbidden)
		return
	} else if err != nil {
		logger.WithFields(log.Fields{
			"trackmeta": trackMeta,
			"error":     err,
		}).Info("unable to add track metadata")
		http.Error(w, "internal server error occurred", http.StatusInternalServerError)
		return
	}

	// Send the ticker information that we just added a track
	server.ticker.Reporter() <- trackMeta.Timestamp

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

// trackGetAllHandler returns all ids of registered igc files
func (server *Server) trackGetAllHandler(w http.ResponseWriter, r *http.Request) {
	logger := newReqLogger(r)

	logger.Info("processing request to get all track ids")

	ids, err := server.tracks.GetAllIDs()
	if err != nil {
		logger.WithField("error", err).Error("unable to respond to request of all IDs")
		http.Error(w, "internal server error occurred", http.StatusInternalServerError)
		return
	}
	logger.WithField("ids", ids).Info("responding to request with all ids")
	json.NewEncoder(w).Encode(ids)
}

// trackGetHandler should return the fields of a specific id
func (server *Server) trackGetHandler(w http.ResponseWriter, r *http.Request) {
	logger := newReqLogger(r)

	logger.Info("processing request to get specific track")

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
	meta, err := server.tracks.Get(TrackID(id))
	if err == ErrTrackNotFound {
		idlog.Info("unable to find metadata of id")
		http.Error(w, "content not found", http.StatusNotFound)
		return
	} else if err != nil {
		idlog.WithField("error", err).Info("error when getting metadata of id")
		http.Error(w, "internal server error occurred", http.StatusInternalServerError)
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

	logger.Info("processing request to get field of specific track")

	vars := mux.Vars(r)
	idStr, _ := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.WithField("id", idStr).Info("id must be a valid number")
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	field, _ := vars["field"]
	idlog := logger.WithField("id", id)

	meta, err := server.tracks.Get(TrackID(id))
	if err == ErrTrackNotFound {
		idlog.Info("unable to find metadata of id")
		http.Error(w, "content not found", http.StatusNotFound)
		return
	} else if err != nil {
		idlog.WithField("error", err).Info("error when getting metadata of id")
		http.Error(w, "internal server error occurred", http.StatusInternalServerError)
		return
	}

	flog := idlog.WithField("field", field)
	switch field {
	case "H_date":
		flog.Info("responding with track date")
		io.WriteString(w, meta.Date.Format(time.RFC3339))
	case "pilot":
		flog.Info("responding with track pilot")
		io.WriteString(w, meta.Pilot)
	case "glider":
		flog.Info("responding with track glider")
		io.WriteString(w, meta.Glider)
	case "glider_id":
		flog.Info("responding with track glider id")
		io.WriteString(w, meta.GliderID)
	case "track_length":
		flog.Info("responding with track length")
		io.WriteString(w, strconv.FormatFloat(meta.TrackLength, 'f', -1, 64))
	case "track_src_url":
		flog.Info("responding with track src url")
		io.WriteString(w, meta.TrackSrcURL)
	default:
		flog.Info("unable to find field of metadata")
		http.Error(w, "invalid field", http.StatusBadRequest)
	}
}
