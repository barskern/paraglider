package igcserver

import (
	"math/rand"
	"sync"
	"testing"
)

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
