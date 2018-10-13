package main

import (
	"github.com/marni/goigc"
	"math/rand"
	"sync"
	"time"
)

// A unique id for a track
type trackID string

// NewTrackID creates a new unique track ID
func NewTrackID() string {
	id := [8]byte{}
	rand.Read(id[:])
	return string(id[:])
}

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

// calcTotalDistance returns the total distance between the points in order
func calcTotalDistance(points []igc.Point) (trackLength float64) {
	for i, p := range points {
		trackLength += p.Distance(points[i+1])
	}
	return
}

// TrackMetaFrom converts a igc.Track into a TrackMeta struct
func TrackMetaFrom(track igc.Track) TrackMeta {
	return TrackMeta{
		track.Date,
		track.Pilot,
		track.GliderType,
		track.GliderID,
		calcTotalDistance(track.Points),
	}
}

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
