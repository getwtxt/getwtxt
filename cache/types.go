package cache

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

// Data from user's twtxt.txt
type Data struct {
	nick    string
	date    time.Time
	apidate []byte
	status  StatusMap
}

// StatusMap holds the statuses posted by a given user. A standard
// time.Time value is used as the key, with the status as a string.
type StatusMap map[time.Time]string

// StatusMapSlice is a slice of StatusMaps. Useful for sorting the
// output of queries.
type StatusMapSlice []StatusMap

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
