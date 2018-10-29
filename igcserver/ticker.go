package igcserver

import (
	"errors"
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
