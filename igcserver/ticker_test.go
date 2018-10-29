package igcserver

import (
	"time"
)

// TickerDummy is simple ticker instance for testing
type TickerDummy struct {
	latest    *time.Time
	publisher chan *time.Time
	reporter  chan time.Time
}

// NewTickerDummy creates a new simple ticker
func NewTickerDummy(buf int) TickerDummy {
	reporter := make(chan time.Time, buf)
	publisher := make(chan *time.Time)
	ticker := TickerDummy{
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
			case latest := <-reporter:
				ticker.latest = &latest
			// A user asked for the latest value so we send it
			case publisher <- ticker.latest:
			}
		}
	}()

	return ticker
}

// Reporter returns a channel which expects to have timestamps sent to it and
// it will listen to it and keep the current timestamp updated accordingly
func (t *TickerDummy) Reporter(latest time.Time) {
	t.reporter <- latest
}

// Latest returns a channel which expects send the current latest timestamp on
// a request
func (t *TickerDummy) Latest() *time.Time {
	return <-t.publisher
}

// GetReport returns a report of the oldest registered timestamps with the given limit
func (t *TickerDummy) GetReport(limit int) (rep TickerReport, err error) {
	rep, err = t.GetReportAfter(time.Unix(0, 0), limit)
	return
}

// GetReportAfter returns a report after a specified time with the given limit
func (t *TickerDummy) GetReportAfter(timestamp time.Time, limit int) (rep TickerReport, err error) {
	err = ErrNoTracksFound
	return
}
