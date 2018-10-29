package igcserver

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"time"
)

var (
	// ErrNoTracksFound symbolizes that there are no tracks to report the
	// timestamp on
	ErrNoTracksFound = errors.New("no tracks meeting criteria found")
)

// Ticker is a generic interface for any type which can act as a ticker
type Ticker interface {
	Latest() <-chan *time.Time
	Reporter() chan<- time.Time
	GetReport(limit int) (TickerReport, error)
	GetReportAfter(timestamp time.Time, limit int) (TickerReport, error)
}

// TickerReport encompasses the report the ticker provides to the user
//
// {
// "t_latest": <latest added timestamp>,
// "t_start": <the first timestamp of the added track>,
// "t_stop": <the last timestamp of the added track>,
// "tracks": [<id1>, <id2>, ...],
// "processing": <time in ms of how long it took to process the request>
// }
type TickerReport struct {
	Latest     time.Time     `json:"t_latest"`
	Start      time.Time     `json:"t_start"`
	End        time.Time     `json:"t_stop"`
	Tracks     []TrackID     `json:"tracks"`
	Processing time.Duration `json:"processing"`
}

// ---------- //
// TICKER API //
// ---------- //

func (server *Server) tickerHandler(w http.ResponseWriter, r *http.Request) {
	logger := newReqLogger(r)

	logger.Info("processing request to get ticker report")

	report, err := server.ticker.GetReport(5)
	if err == ErrNoTracksFound {
		logger.WithField("error", err).Info("no tracks registered")
		http.Error(w, "content not found", http.StatusNotFound)
		return
	} else if err != nil {
		logger.WithField("error", err).Info("unable to build ticker report")
		http.Error(w, "internal server error occurred", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(report)
}

func (server *Server) tickerAfterHandler(w http.ResponseWriter, r *http.Request) {
	logger := newReqLogger(r)

	logger.Info("processing request to get ticker report after timestamp")

	vars := mux.Vars(r)
	timestampStr, _ := vars["timestamp"]
	timestamp, err := time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		logger.WithField("error", err).Info("unable to parse as timestamp")
		http.Error(w, "invalid timestamp", http.StatusBadRequest)
		return
	}

	report, err := server.ticker.GetReportAfter(timestamp, 5)
	if err == ErrNoTracksFound {
		logger.WithField("error", err).Info("no tracks registered")
		http.Error(w, "content not found", http.StatusNotFound)
		return
	} else if err != nil {
		logger.WithField("error", err).Info("unable to build ticker report")
		http.Error(w, "internal server error occurred", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(report)
}

func (server *Server) tickerLatestHandler(w http.ResponseWriter, r *http.Request) {
	logger := newReqLogger(r)

	logger.Info("processing request to get latest ticker timestamp")

	latest := <-server.ticker.Latest()
	if latest == nil {
		logger.Info("latest timestamp of request not set")
		http.Error(w, "content not found", http.StatusNotFound)
		return
	}
	logger.WithField("latest", latest).Info("responding with latest timestamp")
	io.WriteString(w, latest.Format(time.RFC3339))
}
