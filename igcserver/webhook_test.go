package igcserver

import (
	"math/rand"
	"sync"
	"testing"
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
	trigger := make(chan bool, 3)
	go func() {
		<-trigger
	}()
	return WebhooksMap{sync.RWMutex{}, make(map[WebhookID]WebhookInfo), trigger}
}

// Trigger returns a channel to trigger webhooks based on the number of tracks
func (db *WebhooksMap) Trigger() {
	db.trigger <- true
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

// Test that all returned ids from 'Append' are found when using 'Get'
func TestWebhookDuplicate(t *testing.T) {
	webhook := WebhookInfo{
		ID: NewWebhookID([]byte("not-unique")),
	}

	webhooks := NewWebhooksMap()

	var err error
	err = webhooks.Append(webhook)
	if err != nil {
		t.Fatalf("unable to add webhook: %s", err)
	}
	err = webhooks.Append(webhook)
	if err == nil {
		t.Fatalf("duplicate webhook id should be rejected")
	}
}

// Test that all returned ids from 'Append' are found when using 'Get' when
// running multiple goroutines
func TestWebhooksGetConcurr(t *testing.T) {
	const hookCount = 10
	var pureHooks [hookCount]WebhookInfo
	buf := make([]byte, 15)
	for i := 0; i < hookCount; i++ {
		rand.Read(buf)
		pureHooks[i] = WebhookInfo{
			ID: NewWebhookID(buf),
		}
	}

	webhooks := NewWebhooksMap()
	var ids [hookCount]WebhookID
	var err error
	for i, v := range pureHooks {
		ids[i] = v.ID
		err = webhooks.Append(v)
		if err != nil {
			t.Fatalf("unable to add hook: %s", err)
		}
	}

	var wg sync.WaitGroup
	for _, pureID := range ids {
		wg.Add(1)
		go func(webhooks *WebhooksMap, id WebhookID) {
			if _, err := webhooks.Get(id); err != nil {
				t.Fatalf("didn't find id '%d' in result of 'Get'", id)
			}
			wg.Done()
		}(&webhooks, pureID)
	}
	wg.Wait()
}

// Test that deleted webhooks are removed
func TestWebhookDeletion(t *testing.T) {
	const hookCount = 10
	var pureHooks [hookCount]WebhookInfo
	buf := make([]byte, 15)
	for i := 0; i < hookCount; i++ {
		rand.Read(buf)
		pureHooks[i] = WebhookInfo{
			ID: NewWebhookID(buf),
		}
	}

	webhooks := NewWebhooksMap()
	var ids [hookCount]WebhookID
	var err error
	for i, v := range pureHooks {
		ids[i] = v.ID
		err = webhooks.Append(v)
		if err != nil {
			t.Fatalf("unable to add hook: %s", err)
		}
	}

	var webhook WebhookInfo
	if webhook, err = webhooks.Delete(ids[0]); err != nil {
		t.Fatalf("unable to delete webhook which was just inserted")
	}

	if webhook.ID != ids[0] {
		t.Fatalf("id of removed webhook was not equal to the id specifid when deleting")
	}
}
