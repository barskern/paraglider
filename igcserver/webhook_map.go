package igcserver

import (
	"sync"
)

// WebhooksMap contains a map to many WebhookInfo objects which are protected
// by a RWMutex and indexed by a unique id
type WebhooksMap struct {
	sync.RWMutex
	data map[WebhookID]WebhookInfo
}

// NewWebhooksMap creates a new mutex and mapping from ID to WebhookInfo
func NewWebhooksMap() WebhooksMap {
	return WebhooksMap{sync.RWMutex{}, make(map[WebhookID]WebhookInfo)}
}

// Get fetches the webhook of a specific id if it exists
func (webhooks *WebhooksMap) Get(id WebhookID) (webhook WebhookInfo, err error) {
	webhooks.RLock()
	defer webhooks.RUnlock()
	webhook, ok := webhooks.data[id]
	if !ok {
		err = ErrWebhookNotFound
	}
	return
}

// Append appends a webhook and returns the given id
func (webhooks *WebhooksMap) Append(webhook WebhookInfo) (err error) {
	webhooks.Lock()
	defer webhooks.Unlock()
	if _, exists := webhooks.data[webhook.ID]; exists {
		err = ErrWebhookAlreadyExists
	} else {
		webhooks.data[webhook.ID] = webhook
	}
	return
}

// Delete removes a webhook
func (webhooks *WebhooksMap) Delete(id WebhookID) (webhook WebhookInfo, err error) {
	webhooks.RLock()
	defer webhooks.RUnlock()
	webhook, ok := webhooks.data[id]
	if ok {
		delete(webhooks.data, id)
	} else {
		err = ErrWebhookNotFound
	}
	return
}
