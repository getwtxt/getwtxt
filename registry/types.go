package registry

import (
	"sync"
	"time"
)

// Indexer allows for other uses of the Index functions
type Indexer interface {
	AddUser(string, string)
	DelUser(string)
}

// UserIndex provides an index of users by URL
type UserIndex map[string]*Data

// Data on each user. `Nick` is the specified nickname. `Date` is the
// time.Time of the user's submission to the registry. `APIdate` is the
// RFC3339-formatted date/time of the user's submission. `Status` is a
// TimeMap containing the user's statuses.
type Data struct {
	Nick    string
	Date    time.Time
	APIdate []byte
	Status  TimeMap
}

// TimeMap holds extracted and processed user data as a
// string. A standard time.Time value is used as the key.
type TimeMap map[time.Time]string

// TimeMapSlice is a slice of TimeMap. Useful for sorting the
// output of queries.
type TimeMapSlice []TimeMap

// Mutex to control access to the User Index.
var imutex = sync.RWMutex{}

// TimeSlice is used for sorting by timestamp.
type TimeSlice []time.Time

// Len returns the length of the slice to be sorted.
// This helps satisfy sort.Interface with respect to TimeSlice.
func (t TimeSlice) Len() int {
	return len(t)
}

// Less returns true if the timestamp at index i is before the
// timestamp at index j in TimeSlice.
// This helps satisfy sort.Interface with respect to TimeSlice.
func (t TimeSlice) Less(i, j int) bool {
	return t[i].Before(t[j])
}

// Swap transposes the timestampss at the two given indices.
// This helps satisfy sort.Interface with respect to TimeSlice.
func (t TimeSlice) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}
