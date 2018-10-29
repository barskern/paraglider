package igcserver

import (
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
func (metas *TrackMetasDB) Get(id TrackID) (meta TrackMeta, err error) {
	err = metas.Find(bson.M{"id": id}).One(&meta)
	if err == mgo.ErrNotFound {
		err = ErrTrackNotFound
	}
	return
}

// Append appends a track meta and returns the given id
func (metas *TrackMetasDB) Append(meta TrackMeta) (err error) {
	_, err = metas.Get(meta.ID)
	if err == ErrTrackNotFound {
		err = metas.Insert(meta)
	} else if err == nil {
		err = ErrTrackAlreadyExists
	}
	return
}

// GetAllIDs fetches all the stored ids
func (metas *TrackMetasDB) GetAllIDs() (ids []TrackID, err error) {
	var trackMetas []TrackMeta
	err = metas.Find(nil).All(&trackMetas)
	if err == nil {
		ids = make([]TrackID, len(trackMetas))
		for i, v := range trackMetas {
			ids[i] = v.ID
		}
	}
	return
}
