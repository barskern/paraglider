package igcserver

import (
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"time"
)

// TickerDB is a database-aware ticker instance
type TickerDB struct {
	session   *mgo.Session
	latest    *time.Time
	publisher chan *time.Time
	reporter  chan time.Time
}

// NewTickerDB creates a new database-aware ticker instance
func NewTickerDB(session *mgo.Session, buf int) TickerDB {
	reporter := make(chan time.Time, buf)
	publisher := make(chan *time.Time)
	ticker := TickerDB{
		session,
		nil,
		publisher,
		reporter,
	}

	// Handle all requests and responses to get latest ticker value. Doesn't
	// need a mutex because all accesses to ticker.latest is done in the
	// following goroutine
	go func() {
		for {
			select {
			// We received a new latest value
			case *ticker.latest = <-reporter:
			// A user asked for the latest value so we send it
			case publisher <- ticker.latest:
			}
		}
	}()

	return ticker
}

// Reporter returns a channel which expects to have timestamps sent to it and
// it will listen to it and keep the current timestamp updated accordingly
func (t *TickerDB) Reporter() chan<- time.Time {
	return t.reporter
}

// Latest returns a channel which expects send the current latest timestamp on
// a request
func (t *TickerDB) Latest() <-chan *time.Time {
	return t.publisher
}

// GetReport returns a report of the oldest registered timestamps with the given limit
func (t *TickerDB) GetReport(limit int) (rep TickerReport, err error) {
	rep, err = t.GetReportAfter(time.Unix(0, 0), limit)
	return
}

// GetReportAfter returns a report after a specified time with the given limit
func (t *TickerDB) GetReportAfter(timestamp time.Time, limit int) (rep TickerReport, err error) {
	start := time.Now()

	conn := t.session.Copy()
	defer conn.Close()
	tracks := conn.DB("").C(trackCollection)

	var trackMetas []TrackMeta
	err = tracks.
		Find(bson.M{"timestamp": bson.M{"$gt": timestamp}}).
		Limit(limit).
		Sort("timestamp").
		All(trackMetas)

	if err == nil {
		if len(trackMetas) < 1 {
			err = ErrNoTracksFound
		} else {
			latest := <-t.Latest()
			firststamp := trackMetas[0].Timestamp
			laststamp := trackMetas[len(trackMetas)-1].Timestamp
			ids := make([]TrackID, len(trackMetas))
			for i, meta := range trackMetas {
				ids[i] = meta.ID
			}
			processing := time.Since(start)
			rep = TickerReport{
				*latest,
				firststamp,
				laststamp,
				ids,
				processing,
			}
		}
	}
	return
}
