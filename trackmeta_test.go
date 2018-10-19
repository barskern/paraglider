package main

import (
	"sync"
	"testing"
)

// Test that the generated ids are unique
func TestTrackIDGeneration(t *testing.T) {
	const idCount = 100
	var ids [idCount]TrackID
	for i := 0; i < idCount; i++ {
		newID := NewTrackID()

		for j := 0; j < i; j++ {
			if ids[j] == newID {
				t.Fatalf("generated a duplicated id in '%d' attempts", idCount)
			}
		}

		ids[i] = newID
	}
}

// Test that all returned ids from 'Append' are found when using 'Get'
func TestTrackMetasGet(t *testing.T) {
	const metaCount = 10
	pureMetas := make([]TrackMeta, metaCount)

	metas := NewTrackMetas()

	var ids [metaCount]TrackID
	for i, v := range pureMetas {
		ids[i] = metas.Append(v)
	}

	for _, pureID := range ids {
		if _, ok := metas.Get(pureID); !ok {
			t.Fatalf("didn't find id '%s' in result of 'GetAllIDs'", pureID)
		}
	}
}

// Test that all returned ids from 'Append' are found when using 'Get' when
// running multiple goroutines
func TestTrackMetasGetConcurr(t *testing.T) {
	const metaCount = 10
	pureMetas := make([]TrackMeta, metaCount)

	metas := NewTrackMetas()

	var ids [metaCount]TrackID
	for i, v := range pureMetas {
		ids[i] = metas.Append(v)
	}

	var wg sync.WaitGroup
	for _, pureID := range ids {
		wg.Add(1)
		go func(metas *TrackMetas, id TrackID) {
			if _, ok := metas.Get(id); !ok {
				t.Fatalf("didn't find id '%s' in result of 'GetAllIDs'", id)
			}
			wg.Done()
		}(&metas, pureID)
	}
	wg.Wait()
}

// Test that all returned ids from 'Append' are found in the output of 'GetAllIDs'
func TestTrackMetasGetAllIDs(t *testing.T) {
	const metaCount = 10
	pureMetas := make([]TrackMeta, metaCount)

	metas := NewTrackMetas()

	var ids [metaCount]TrackID
	for i, v := range pureMetas {
		ids[i] = metas.Append(v)
	}

	for _, pureID := range ids {
		found := false
		for _, metaID := range metas.GetAllIDs() {
			if pureID == metaID {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("didn't find id '%s' in result of 'GetAllIDs'", pureID)
		}
	}
}
