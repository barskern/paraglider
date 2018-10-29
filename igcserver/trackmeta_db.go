package igcserver

import (
	"errors"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// TrackMetasDB contains a map to many TrackMeta objects which are protected
// by a RWMutex and indexed by a unique id
type TrackMetasDB struct {
	*mgo.Collection
}

// NewTrackMetasDB creates a new mutex and mapping from ID to TrackMeta
func NewTrackMetasDB(collection *mgo.Collection) TrackMetasDB {
	return TrackMetasDB{
		collection,
	}
}

// Get fetches the track meta of a specific id if it exists
func (metas *TrackMetasDB) Get(id TrackID) (TrackMeta, bool, error) {
	var meta TrackMeta
	err := metas.Find(bson.M{"id": id}).One(&meta)
	if err == mgo.ErrNotFound {
		return TrackMeta{}, false, nil
	}
	if err != nil {
		return TrackMeta{}, false, err
	}
	return meta, true, nil
}

// Append appends a track meta and returns the given id
func (metas *TrackMetasDB) Append(meta TrackMeta) error {
	_, exists, err := metas.Get(meta.ID)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("track already exists")
	}
	return metas.Insert(meta)
}

// GetAllIDs fetches all the stored ids
func (metas *TrackMetasDB) GetAllIDs() ([]TrackID, error) {
	var trackMetas []TrackMeta
	err := metas.Find(nil).All(&trackMetas)
	if err != nil {
		return []TrackID{}, err
	}
	ids := make([]TrackID, len(trackMetas))
	for i, v := range trackMetas {
		ids[i] = v.ID
	}
	return ids, nil
}
