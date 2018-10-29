package igcserver

import (
	"math/rand"
	"sync"
	"testing"
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

// Test that all returned ids from 'Append' are found when using 'Get'
func TestTrackMetaDuplicate(t *testing.T) {
	meta := TrackMeta{
		ID: NewTrackID([]byte("not-unique")),
	}

	metas := NewTrackMetasMap()

	var err error
	err = metas.Append(meta)
	if err != nil {
		t.Fatalf("unable to add metadata: %s", err)
	}
	err = metas.Append(meta)
	if err == nil {
		t.Fatalf("same track meta duplicate track ids should be rejected")
	}
}

// Test that all returned ids from 'Append' are found when using 'Get'
func TestTrackMetasGet(t *testing.T) {
	const metaCount = 10
	var pureMetas [metaCount]TrackMeta
	buf := make([]byte, 15)
	for i := 0; i < metaCount; i++ {
		rand.Read(buf)
		pureMetas[i] = TrackMeta{
			ID: NewTrackID(buf),
		}
	}

	metas := NewTrackMetasMap()

	var ids [metaCount]TrackID
	var err error
	for i, v := range pureMetas {
		ids[i] = v.ID
		err = metas.Append(v)
		if err != nil {
			t.Fatalf("unable to add metadata: %s", err)
			continue
		}
	}

	for _, pureID := range ids {
		if _, err := metas.Get(pureID); err != nil {
			t.Fatalf("didn't find id '%d' in result of 'GetAllIDs'", pureID)
		}
	}
}

// Test that all returned ids from 'Append' are found when using 'Get' when
// running multiple goroutines
func TestTrackMetasGetConcurr(t *testing.T) {
	const metaCount = 10
	var pureMetas [metaCount]TrackMeta
	buf := make([]byte, 15)
	for i := 0; i < metaCount; i++ {
		rand.Read(buf)
		pureMetas[i] = TrackMeta{
			ID: NewTrackID(buf),
		}
	}

	metas := NewTrackMetasMap()
	var ids [metaCount]TrackID
	var err error
	for i, v := range pureMetas {
		ids[i] = v.ID
		err = metas.Append(v)
		if err != nil {
			t.Fatalf("unable to add metadata: %s", err)
		}
	}

	var wg sync.WaitGroup
	for _, pureID := range ids {
		wg.Add(1)
		go func(metas *TrackMetasMap, id TrackID) {
			if _, err := metas.Get(id); err != nil {
				t.Fatalf("didn't find id '%d' in result of 'GetAllIDs'", id)
			}
			wg.Done()
		}(&metas, pureID)
	}
	wg.Wait()
}

// Test that all returned ids from 'Append' are found in the output of 'GetAllIDs'
func TestTrackMetasGetAllIDs(t *testing.T) {
	const metaCount = 10
	var pureMetas [metaCount]TrackMeta
	buf := make([]byte, 15)
	for i := 0; i < metaCount; i++ {
		rand.Read(buf)
		pureMetas[i] = TrackMeta{
			ID: NewTrackID(buf),
		}
	}

	metas := NewTrackMetasMap()
	var ids [metaCount]TrackID
	var err error
	for i, v := range pureMetas {
		ids[i] = v.ID
		err = metas.Append(v)
		if err != nil {
			t.Fatalf("unable to add metadata: %s", err)
		}
	}

	for _, pureID := range ids {
		found := false
		metaIDs, _ := metas.GetAllIDs()
		for _, metaID := range metaIDs {
			if pureID == metaID {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("didn't find id '%d' in result of 'GetAllIDs'", pureID)
		}
	}
}
