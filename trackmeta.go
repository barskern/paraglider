package main

import (
	"sync"
	"time"
)

// A unique id for a track
// TODO give a meaningful type, perhaps a generated uuid
type trackID int

// TrackMeta contains a subset of metainformation about a igc-track
//
// ```json
// {
//   "H_date": <date from File Header, H-record>,
//   "pilot": <pilot>,
//   "glider": <glider>,
//   "glider_id": <glider_id>,
//   "track_length": <calculated total track length>
// }
// ```
type TrackMeta struct {
	Date        time.Time `json:"H_date"`
	Pilot       string    `json:"pilot"`
	Gilder      string    `json:"glider"`
	GliderID    string    `json:"glider_id"`
	TrackLength float64   `json:"track_length"`
}

// TODO make function which converts a igc.Track into a TrackMeta struct.
// See https://godoc.org/github.com/marni/goigc#Track

// TrackMetas contains a map to many TrackMeta objects which are protected
// by a RWMutex and indexed by a unique id
type TrackMetas struct {
	mu   sync.RWMutex
	data map[trackID]TrackMeta
}

// TODO implement a read and write function which returns the map and the mutex
// after locking

// NewTrackMetas creates a new mutex and mapping from ID to TrackMeta
func NewTrackMetas() TrackMetas {
	return TrackMetas{sync.RWMutex{}, make(map[trackID]TrackMeta)}
}
