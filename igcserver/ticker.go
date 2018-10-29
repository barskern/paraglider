package igcserver

import (
	"time"
)

// Ticker is a generic interface for any type which can act as a ticker
type Ticker interface {
	ReportLatest() <-chan time.Time
	GetReport() (TickerReport, error)
	GetReportAfter(time.Time) (TickerReport, error)
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
