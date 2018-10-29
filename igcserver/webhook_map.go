package igcserver

import (
	"sync"
)

// WebhooksMap contains a map to many WebhookInfo objects which are protected
// by a RWMutex and indexed by a unique id
type WebhooksMap struct {
	sync.RWMutex
	data    map[WebhookID]WebhookInfo
	trigger chan bool
}

// NewWebhooksMap creates a new mutex and mapping from ID to WebhookInfo
func NewWebhooksMap() WebhooksMap {
	return WebhooksMap{sync.RWMutex{}, make(map[WebhookID]WebhookInfo), make(chan bool)}
}

// Trigger returns a channel to trigger webhooks based on the number of tracks
func (db *WebhooksMap) Trigger() chan<- bool {
	return db.trigger
}

// Get fetches the webhook of a specific id if it exists
func (db *WebhooksMap) Get(id WebhookID) (webhook WebhookInfo, err error) {
	db.RLock()
	defer db.RUnlock()
	webhook, ok := db.data[id]
	if !ok {
		err = ErrWebhookNotFound
	}
	return
}

// Append appends a webhook and returns the given id
func (db *WebhooksMap) Append(webhook WebhookInfo) (err error) {
	db.Lock()
	defer db.Unlock()
	if _, exists := db.data[webhook.ID]; exists {
		err = ErrWebhookAlreadyExists
	} else {
		db.data[webhook.ID] = webhook
	}
	return
}

// Delete removes a webhook
func (db *WebhooksMap) Delete(id WebhookID) (webhook WebhookInfo, err error) {
	db.RLock()
	defer db.RUnlock()
	webhook, ok := db.data[id]
	if ok {
		delete(db.data, id)
	} else {
		err = ErrWebhookNotFound
	}
	return
}
