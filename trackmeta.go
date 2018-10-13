package main

import (
	"github.com/marni/goigc"
	"math/rand"
	"sync"
	"time"
)

// TrackID is a unique id for a track
type TrackID string

// NewTrackID creates a new unique track ID
func NewTrackID() TrackID {
	id := [8]byte{}
	rand.Read(id[:])
	return TrackID(string(id[:]))
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
	data map[TrackID]TrackMeta
}

// Get fetches the track meta of a specific id if it exists
func (metas *TrackMetas) Get(id TrackID) (TrackMeta, bool) {
	metas.mu.RLock()
	defer metas.mu.RUnlock()
	v, ok := metas.data[id]
	return v, ok
}

// Append appends a track meta and returns the given id
func (metas *TrackMetas) Append(meta TrackMeta) TrackID {
	// TODO perhaps check that id doesnt already exist?
	id := NewTrackID()
	metas.mu.Lock()
	defer metas.mu.Unlock()
	metas.data[id] = meta
	return id
}

// NewTrackMetas creates a new mutex and mapping from ID to TrackMeta
func NewTrackMetas() TrackMetas {
	return TrackMetas{sync.RWMutex{}, make(map[TrackID]TrackMeta)}
}
