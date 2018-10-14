package main

import (
	"encoding/base64"
	"github.com/marni/goigc"
	"math/rand"
	"sync"
	"time"
)

// TrackID is a unique id for a track
type TrackID string

// NewTrackID creates a new unique track ID
func NewTrackID() TrackID {
	var id [6]byte
	rand.Read(id[:])
	// Encode random bytes as a base64 string so that it only uses valid ascii
	// characters (this is to make ids "pretty" in the url)
	return TrackID(base64.StdEncoding.EncodeToString(id[:]))
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
	Glider      string    `json:"glider"`
	GliderID    string    `json:"glider_id"`
	TrackLength float64   `json:"track_length"`
}

// calcTotalDistance returns the total distance between the points in order
func calcTotalDistance(points []igc.Point) (trackLength float64) {
	for i := 0; i+1 < len(points); i++ {
		trackLength += points[i].Distance(points[i+1])
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
	sync.RWMutex
	data map[TrackID]TrackMeta
}

// NewTrackMetas creates a new mutex and mapping from ID to TrackMeta
func NewTrackMetas() TrackMetas {
	return TrackMetas{sync.RWMutex{}, make(map[TrackID]TrackMeta)}
}

// Get fetches the track meta of a specific id if it exists
func (metas *TrackMetas) Get(id TrackID) (TrackMeta, bool) {
	metas.RLock()
	defer metas.RUnlock()
	v, ok := metas.data[id]
	return v, ok
}

// Append appends a track meta and returns the given id
func (metas *TrackMetas) Append(meta TrackMeta) TrackID {
	// TODO perhaps check that id doesnt already exist?
	id := NewTrackID()
	metas.Lock()
	defer metas.Unlock()
	metas.data[id] = meta
	return id
}

// GetAllIDs fetches all the stored ids
func (metas *TrackMetas) GetAllIDs() []TrackID {
	metas.RLock()
	defer metas.RUnlock()
	keys := make([]TrackID, 0, len(metas.data))
	for k := range metas.data {
		keys = append(keys, k)
	}
	return keys
}
