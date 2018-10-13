package main

import "testing"

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
