package igcserver

import (
	"encoding/json"
	"github.com/barskern/paragliding/isodur"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

// Server distributes request to a pool of worker gorutines
type Server struct {
	startupTime time.Time
	httpClient  *http.Client
	router      *mux.Router
	ticker      Ticker
	tracks      TrackMetas
	webhooks    Webhooks
}

// NewServer creates a new server which handles requests to the igc api
func NewServer(httpClient *http.Client, trackMetas TrackMetas, ticker Ticker, webhooks Webhooks) (srv Server) {
	srv = Server{
		time.Now(),
		httpClient,
		mux.NewRouter(),
		ticker,
		trackMetas,
		webhooks,
	}

	srv.router.Use(loggingMiddleware)

	// Webhook API
	srv.router.HandleFunc("/webhook/new_track", srv.webhookRegHandler).Methods(http.MethodPost)
	srv.router.HandleFunc("/webhook/new_track/{webhookID}", srv.webhookGetHandler).Methods(http.MethodGet)
	srv.router.HandleFunc("/webhook/new_track/{webhookID}", srv.webhookDeleteHandler).Methods(http.MethodDelete)

	// Ticker API
	srv.router.HandleFunc("/ticker", srv.tickerHandler).Methods(http.MethodGet)
	srv.router.HandleFunc("/ticker/latest", srv.tickerLatestHandler).Methods(http.MethodGet)
	srv.router.HandleFunc("/ticker/{timestamp}", srv.tickerAfterHandler).Methods(http.MethodGet)

	// Igc track API
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

	logger.Info("processing request to get metadata")

	metadata := map[string]interface{}{
		"uptime":  isodur.FormatAsISO8601(time.Since(server.startupTime)),
		"info":    "Service for Paragliding tracks.",
		"version": "v1",
	}

	logger.WithFields(log.Fields(metadata)).Info("responding with metadata")

	// Encode metadata as a JSON object
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metadata)
}
