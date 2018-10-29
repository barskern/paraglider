package igcserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"hash/fnv"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

var (
	// ErrWebhookNotFound is returned if a request did not result in a webhook
	ErrWebhookNotFound = errors.New("webhook not found")

	// ErrWebhookAlreadyExists is returned to request to add a webhook which
	// already exists
	ErrWebhookAlreadyExists = errors.New("webhook already exists")
)

// Webhooks is a interface for all storages containing WebhookInfo
type Webhooks interface {
	Trigger() chan<- bool
	Get(id WebhookID) (WebhookInfo, error)
	Append(webhook WebhookInfo) error
	Delete(id WebhookID) (WebhookInfo, error)
}

// WebhookInfo contains information about a webhook
type WebhookInfo struct {
	ID            WebhookID `json:"-" bson:"id"`
	URLstr        string    `json:"webhookURL" bson:"webhookURL"`
	TriggerRate   uint      `json:"minTriggerValue" bson:"minTriggerValue"`
	LastTriggered time.Time `json:"-" bson:"lastTriggered"`
}

// WebhookID is a unique id for a track
type WebhookID uint32

// NewWebhookID creates a new unique track ID
func NewWebhookID(v []byte) WebhookID {
	hasher := fnv.New32()
	hasher.Write(v)
	return WebhookID(hasher.Sum32())
}

// ----------- //
// WEBHOOK API //
// ----------- //

func (server *Server) webhookRegHandler(w http.ResponseWriter, r *http.Request) {
	logger := newReqLogger(r)

	logger.Info("processing request to register webhook")

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	webhook := WebhookInfo{TriggerRate: 1}
	if err := dec.Decode(&webhook); err != nil {
		logger.WithField("error", err).Info("unable to decode request body")
		http.Error(w, "invalid json object", http.StatusBadRequest)
		return
	}
	reqURL, err := url.Parse(webhook.URLstr)
	if err != nil {
		logger.WithField("error", err).Info("unable to parse url")
		http.Error(w, "invalid url", http.StatusBadRequest)
		return
	}
	if webhook.TriggerRate < 1 {
		logger.WithField("error", err).Info("invalid trigger value")
		http.Error(w, "invalid trigger value", http.StatusBadRequest)
		return
	}
	webhook.ID = NewWebhookID([]byte(reqURL.String()))
	err = server.webhooks.Append(webhook)
	if err == ErrWebhookAlreadyExists {
		logger.WithFields(log.Fields{
			"webhook": webhook,
		}).Info("request attempted to add duplicate webhook")
		http.Error(w, "webhook already exists", http.StatusForbidden)
		return
	} else if err != nil {
		logger.WithFields(log.Fields{
			"webhook": webhook,
			"error":   err,
		}).Info("unable to add webhook")
		http.Error(w, "internal server error occurred", http.StatusInternalServerError)
		return
	}
	logger.WithFields(log.Fields{
		"webhook": webhook,
	}).Info("added webhook")

	io.WriteString(w, fmt.Sprintf("%d", webhook.ID))
}

func (server *Server) webhookGetHandler(w http.ResponseWriter, r *http.Request) {
	logger := newReqLogger(r)

	logger.Info("processing request to get webhook")

	vars := mux.Vars(r)
	idStr, _ := vars["webhookID"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.WithField("id", idStr).Info("id must be a valid number")
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	idlog := logger.WithField("id", id)
	webhook, err := server.webhooks.Get(WebhookID(id))
	if err == ErrWebhookNotFound {
		idlog.Info("unable to find metadata of id")
		http.Error(w, "content not found", http.StatusNotFound)
		return
	} else if err != nil {
		idlog.WithField("error", err).Info("error when getting webhook of id")
		http.Error(w, "internal server error occurred", http.StatusInternalServerError)
		return
	}
	logger.WithFields(log.Fields{
		"webhook": webhook,
	}).Info("responding with info about webhook")

	json.NewEncoder(w).Encode(webhook)
}

func (server *Server) webhookDeleteHandler(w http.ResponseWriter, r *http.Request) {
	logger := newReqLogger(r)

	logger.Info("processing request to delete webhook")

	vars := mux.Vars(r)
	idStr, _ := vars["webhookID"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.WithField("id", idStr).Info("id must be a valid number")
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	idlog := logger.WithField("id", id)
	webhook, err := server.webhooks.Delete(WebhookID(id))
	if err == ErrWebhookNotFound {
		idlog.Info("unable to find metadata of id")
		http.Error(w, "content not found", http.StatusNotFound)
		return
	} else if err != nil {
		idlog.WithField("error", err).Info("error when deleting webhook of id")
		http.Error(w, "internal server error occurred", http.StatusInternalServerError)
		return
	}
	idlog.WithFields(log.Fields{
		"webhook": webhook,
	}).Info("responding with info about deleted webhook")

	json.NewEncoder(w).Encode(webhook)
}
