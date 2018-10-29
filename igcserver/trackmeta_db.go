package igcserver

import (
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

const (
	trackCollection = "igctracks"
)

// TrackMetasDB contains a map to many TrackMeta objects which are protected
// by a RWMutex and indexed by a unique id
type TrackMetasDB struct {
	session *mgo.Session
}

// NewTrackMetasDB creates a new mutex and mapping from ID to TrackMeta
func NewTrackMetasDB(session *mgo.Session) TrackMetasDB {
	return TrackMetasDB{
		session,
	}
}

// Get fetches the track meta of a specific id if it exists
func (metas *TrackMetasDB) Get(id TrackID) (meta TrackMeta, err error) {
	conn := metas.session.Copy()
	defer conn.Close()
	tracks := conn.DB("").C(trackCollection)

	err = tracks.Find(bson.M{"id": id}).One(&meta)
	if err == mgo.ErrNotFound {
		err = ErrTrackNotFound
	}
	return
}

// Append appends a track meta and returns the given id
func (metas *TrackMetasDB) Append(meta TrackMeta) (err error) {
	conn := metas.session.Copy()
	defer conn.Close()
	tracks := conn.DB("").C(trackCollection)

	n, err := tracks.Find(bson.M{"id": meta.ID}).Count()
	if err == nil {
		if n == 0 {
			err = tracks.Insert(meta)
		} else if n > 0 {
			err = ErrTrackAlreadyExists
		}
	}
	return
}

// GetAllIDs fetches all the stored ids
func (metas *TrackMetasDB) GetAllIDs() (ids []TrackID, err error) {
	conn := metas.session.Copy()
	defer conn.Close()
	tracks := conn.DB("").C(trackCollection)

	var trackMetas []TrackMeta
	err = tracks.Find(nil).All(&trackMetas)
	if err == nil {
		ids = make([]TrackID, len(trackMetas))
		for i, v := range trackMetas {
			ids[i] = v.ID
		}
	}
	return
}
