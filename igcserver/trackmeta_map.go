package igcserver

import (
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
func (metas *TrackMetasMap) Get(id TrackID) (meta TrackMeta, err error) {
	metas.RLock()
	defer metas.RUnlock()
	meta, ok := metas.data[id]
	if !ok {
		err = ErrTrackNotFound
	}
	return
}

// Append appends a track meta and returns the given id
func (metas *TrackMetasMap) Append(meta TrackMeta) (err error) {
	metas.Lock()
	defer metas.Unlock()
	if _, exists := metas.data[meta.ID]; exists {
		err = ErrTrackAlreadyExists
	} else {
		metas.data[meta.ID] = meta
	}
	return
}

// GetAllIDs fetches all the stored ids
func (metas *TrackMetasMap) GetAllIDs() (ids []TrackID, err error) {
	metas.RLock()
	defer metas.RUnlock()
	ids = make([]TrackID, 0, len(metas.data))
	for id := range metas.data {
		ids = append(ids, id)
	}
	return
}
