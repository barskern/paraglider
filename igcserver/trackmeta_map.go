package igcserver

import (
	"errors"
	"sync"
)

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
func (metas *TrackMetasMap) Append(meta TrackMeta) error {
	metas.Lock()
	defer metas.Unlock()
	if _, exists := metas.data[meta.ID]; exists {
		return errors.New("trackmeta with same url already exists")
	}
	metas.data[meta.ID] = meta
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
