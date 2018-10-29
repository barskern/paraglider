package igcserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

const (
	webhookCollection = "webhooks"
)

// WebhooksDB contains a map to many WebhookInfo objects which are protected
// by a RWMutex and indexed by a unique id
type WebhooksDB struct {
	session    *mgo.Session
	httpClient *http.Client
	trigger    chan bool
}

// DiscordMsg is a webhook message that can be sent to discord
type DiscordMsg struct {
	Content string `json:"content"`
}

// NewDiscordMsg creates a new discord message using a template
func NewDiscordMsg(latest time.Time, ids []TrackID, processing time.Duration) DiscordMsg {
	return DiscordMsg{
		fmt.Sprintf(
			"Latest timestamp is %s and the %d new tracks are: %v (processing: %ds %dms)",
			latest.Format(time.RFC3339),
			len(ids),
			ids,
			int(processing.Seconds()),
			(processing.Nanoseconds()/1000)%1000,
		),
	}
}

// NewWebhooksDB creates a new mutex and mapping from ID to WebhookInfo
func NewWebhooksDB(session *mgo.Session, httpClient *http.Client) WebhooksDB {
	trigger := make(chan bool)

	go func() {
		conn := session.Copy()
		defer conn.Close()
		for {
			<-trigger
			iter := conn.DB("").C(webhookCollection).Find(nil).Iter()
			var webhook WebhookInfo
			for iter.Next(&webhook) {
				log.WithField("webhook", webhook).Info("checking if update is needed for webhook")
				go updateWebhook(session.Copy(), httpClient, webhook)
			}
			iter.Close()
		}
	}()

	return WebhooksDB{
		session,
		httpClient,
		trigger,
	}
}

func updateWebhook(conn *mgo.Session, httpClient *http.Client, webhook WebhookInfo) {
	defer conn.Close()

	start := time.Now()

	var trackMetas []TrackMeta
	err := conn.DB("").C(trackCollection).
		Find(bson.M{"timestamp": bson.M{"$gt": webhook.LastTriggered}}).
		Sort("timestamp").
		All(&trackMetas)

	weblog := log.WithField("webhook", webhook)

	if err != nil {
		weblog.WithField("error", err).Error("unable to get track metas after given timestamp")
	}

	if len(trackMetas) >= int(webhook.TriggerRate) {
		laststamp := trackMetas[len(trackMetas)-1].Timestamp
		ids := make([]TrackID, len(trackMetas))
		for i, meta := range trackMetas {
			ids[i] = meta.ID
		}
		processing := time.Since(start)
		msg := NewDiscordMsg(laststamp, ids, processing)

		b := new(bytes.Buffer)
		json.NewEncoder(b).Encode(msg)

		weblog.WithField("msg", msg).Info("sending update to webhook")
		httpClient.Post(webhook.URLstr, "application/json", b)

		// Update last triggered for current webhook
		webhook.LastTriggered = laststamp
		err = conn.DB("").C(webhookCollection).
			Update(bson.M{"id": webhook.ID}, webhook)
	} else {
		weblog.Info("update not needed for webhook")
	}
}

// Trigger returns a channel to trigger webhooks based on the number of tracks
func (db *WebhooksDB) Trigger() {
	db.trigger <- true
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
