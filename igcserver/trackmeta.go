package igcserver

import (
	"errors"
	"github.com/marni/goigc"
	"hash/fnv"
	"net/url"
	"sync"
	"time"
)

// TrackID is a unique id for a track
type TrackID uint32

// NewTrackID creates a new unique track ID
func NewTrackID(v []byte) TrackID {
	hasher := fnv.New32()
	hasher.Write(v)
	return TrackID(hasher.Sum32())
}

// TrackMeta contains a subset of metainformation about a igc-track
//
// ```json
// {
//   "H_date": <date from File Header, H-record>,
//   "pilot": <pilot>,
//   "glider": <glider>,
//   "glider_id": <glider_id>,
//   "track_length": <calculated total track length>,
//   "track_src_url": <the original URL used to upload the track, ie. the URL
//   used with POST>
// }
// ```
type TrackMeta struct {
	Date        time.Time `json:"H_date"`
	Pilot       string    `json:"pilot"`
	Glider      string    `json:"glider"`
	GliderID    string    `json:"glider_id"`
	TrackLength float64   `json:"track_length"`
	TrackSrcURL string    `json:"track_src_url"`
}

// calcTotalDistance returns the total distance between the points in order
func calcTotalDistance(points []igc.Point) (trackLength float64) {
	for i := 0; i+1 < len(points); i++ {
		trackLength += points[i].Distance(points[i+1])
	}
	return
}

// TrackMetaFrom converts a igc.Track into a TrackMeta struct
func TrackMetaFrom(url url.URL, track igc.Track) TrackMeta {
	return TrackMeta{
		track.Date,
		track.Pilot,
		track.GliderType,
		track.GliderID,
		calcTotalDistance(track.Points),
		url.String(),
	}
}

// MakeTrackMetaEntry makes a TrackMeta object and an id for a given track/url
func MakeTrackMetaEntry(url url.URL, track igc.Track) (id TrackID, meta TrackMeta) {
	return NewTrackID([]byte(url.String())), TrackMetaFrom(url, track)
}

// TrackMetas is a interface for all storages containing TrackMeta
type TrackMetas interface {
	Get(id TrackID) (TrackMeta, bool)
	Append(id TrackID, meta TrackMeta) error
	GetAllIDs() []TrackID
}

// TrackMetasMap contains a map to many TrackMeta objects which are protected
// by a RWMutex and indexed by a unique id
type TrackMetasMap struct {
	sync.RWMutex
	data map[TrackID]TrackMeta
}

// NewTrackMetasMap creates a new mutex and mapping from ID to TrackMeta
func NewTrackMetasMap() TrackMetasMap {
	return TrackMetasMap{sync.RWMutex{}, make(map[TrackID]TrackMeta)}
}

// Get fetches the track meta of a specific id if it exists
func (metas *TrackMetasMap) Get(id TrackID) (TrackMeta, bool) {
	metas.RLock()
	defer metas.RUnlock()
	v, ok := metas.data[id]
	return v, ok
}

// Append appends a track meta and returns the given id
func (metas *TrackMetasMap) Append(id TrackID, meta TrackMeta) error {
	metas.Lock()
	defer metas.Unlock()
	if _, exists := metas.data[id]; exists {
		return errors.New("trackmeta with same url already exists")
	}
	metas.data[id] = meta
	return nil
}

// GetAllIDs fetches all the stored ids
func (metas *TrackMetasMap) GetAllIDs() []TrackID {
	metas.RLock()
	defer metas.RUnlock()
	keys := make([]TrackID, 0, len(metas.data))
	for k := range metas.data {
		keys = append(keys, k)
	}
	return keys
}
