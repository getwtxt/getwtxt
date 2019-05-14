package cache

import (
	"log"
	"sort"
	"strings"
	"time"
)

// NewUserIndex returns a new instance of a user index
func NewUserIndex() *UserIndex {
	return &UserIndex{}
}

// AddUser inserts a new user into the index. The *Data struct
// contains the nickname and the time the user was added.
func (index UserIndex) AddUser(nick string, url string) {
	rfc3339date, err := time.Now().MarshalText()
	if err != nil {
		log.Printf("Error formatting user add time as RFC3339: %v\n", err)
	}
	imutex.Lock()
	index[url] = &Data{nick: nick, date: time.Now(), apidate: rfc3339date}
	imutex.Unlock()
}

// DelUser removes a user from the index completely.
func (index UserIndex) DelUser(url string) {
	imutex.Lock()
	delete(index, url)
	imutex.Unlock()
}

// QueryUser checks the user index for nicknames that contain the
// nickname provided as an argument. Entries are returned sorted
// by the date they were added to the index.
func (index UserIndex) QueryUser(name string) []string {
	var timekey = map[time.Time]string{}
	var sortedkeys TimeSlice
	var users []string
	imutex.RLock()
	for k, v := range index {
		if strings.Contains(v.nick, name) {
			timekey[v.date] = v.nick + "\t" + k + "\t" + string(v.apidate)
			sortedkeys = append(sortedkeys, v.date)
		}
	}
	imutex.RUnlock()
	sort.Sort(sortedkeys)
	for _, e := range sortedkeys {
		users = append(users, timekey[e])
	}

	return users
}

// QueryTag returns all the known statuses that
// contain the provided tag.
func (index UserIndex) QueryTag(tag string) []string {
	var statusmap StatusMapSlice
	i := 0
	imutex.RLock()
	for _, v := range index {
		statusmap[i] = v.FindTag(tag)
		i++
	}
	imutex.RUnlock()

	return statusmap.SortByTime()
}

// FindTag takes a user's tweets and looks for a given tag.
// Returns the tweets with the tag as a []string.
func (userdata *Data) FindTag(tag string) StatusMap {
	var statuses StatusMap
	for k, e := range userdata.status {
		parts := strings.Split(e, "\t")
		statusslice := strings.Split(parts[3], " ")
		for _, v := range statusslice {
			if v[1:] == tag {
				statuses[k] = e
				break
			}
		}
	}

	return statuses
}

// SortByTime returns a string slice of the statuses sorted by time
func (sm StatusMapSlice) SortByTime() []string {
	var tagmap StatusMap
	var times TimeSlice
	var statuses []string
	for _, e := range sm {
		for k, v := range e {
			tagmap[k] = v
		}
	}
	for k := range tagmap {
		times = append(times, k)
	}
	sort.Sort(times)
	for _, e := range times {
		statuses = append(statuses, tagmap[e])
	}

	return statuses
}

// GetStatuses returns the string slice containing a user's statuses
func (index UserIndex) GetStatuses(url string) StatusMap {
	imutex.RLock()
	status := index[url].status
	imutex.RUnlock()
	return status
}
