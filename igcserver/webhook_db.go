package igcserver

import (
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

const (
	webhookCollection = "db"
)

// WebhooksDB contains a map to many WebhookInfo objects which are protected
// by a RWMutex and indexed by a unique id
type WebhooksDB struct {
	session *mgo.Session
}

// NewWebhooksDB creates a new mutex and mapping from ID to WebhookInfo
func NewWebhooksDB(session *mgo.Session) WebhooksDB {
	return WebhooksDB{
		session,
	}
}

// Get fetches the track webhook of a specific id if it exists
func (db *WebhooksDB) Get(id WebhookID) (webhook WebhookInfo, err error) {
	conn := db.session.Copy()
	defer conn.Close()
	webhooks := conn.DB("").C(webhookCollection)

	err = webhooks.Find(bson.M{"id": id}).One(&webhook)
	if err == mgo.ErrNotFound {
		err = ErrWebhookNotFound
	}
	return
}

// Append appends a track webhook and returns the given id
func (db *WebhooksDB) Append(webhook WebhookInfo) (err error) {
	conn := db.session.Copy()
	defer conn.Close()
	webhooks := conn.DB("").C(webhookCollection)

	n, err := webhooks.Find(bson.M{"id": webhook.ID}).Count()
	if err == nil {
		if n == 0 {
			err = webhooks.Insert(webhook)
		} else if n > 0 {
			err = ErrWebhookAlreadyExists
		}
	}
	return
}

// Delete removes a webhook
func (db *WebhooksDB) Delete(id WebhookID) (webhook WebhookInfo, err error) {
	conn := db.session.Copy()
	defer conn.Close()
	webhooks := conn.DB("").C(webhookCollection)

	err = webhooks.Find(bson.M{"id": id}).One(&webhook)
	if err == mgo.ErrNotFound {
		err = ErrWebhookNotFound
	} else if err == nil {
		err = webhooks.Remove(bson.M{"id": id})
	}
	return
}
