package igcserver

import (
	"errors"
	"github.com/marni/goigc"
	"hash/fnv"
	"net/url"
	"time"
)

var (
	// ErrTrackNotFound is returned if a request did not result in a TrackMeta
	ErrTrackNotFound = errors.New("track not found")

	// ErrTrackAlreadyExists is returned to request to add a track which
	// already exists
	ErrTrackAlreadyExists = errors.New("track already exists")
)

// TrackMetas is a interface for all storages containing TrackMeta
type TrackMetas interface {
	Get(id TrackID) (TrackMeta, error)
	Append(meta TrackMeta) error
	GetAllIDs() ([]TrackID, error)
}

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
	ID          TrackID   `json:"-" bson:"id"`
	Timestamp   time.Time `json:"-" bson:"timestamp"`
	Date        time.Time `json:"H_date" bson:"H_date"`
	Pilot       string    `json:"pilot" bson:"pilot"`
	Glider      string    `json:"glider" bson:"glider"`
	GliderID    string    `json:"glider_id" bson:"glider_id"`
	TrackLength float64   `json:"track_length" bson:"track_length"`
	TrackSrcURL string    `json:"track_src_url" bson:"track_src_url"`
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
		NewTrackID([]byte(url.String())),
		time.Now(),
		track.Date,
		track.Pilot,
		track.GliderType,
		track.GliderID,
		calcTotalDistance(track.Points),
		url.String(),
	}
}
