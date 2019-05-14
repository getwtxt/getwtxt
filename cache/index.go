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
	for k, v := range index {
		if strings.Contains(v.nick, name) {
			timekey[v.date] = v.nick + "\t" + k + "\t" + string(v.apidate)
			sortedkeys = append(sortedkeys, v.date)
		}
	}
	sort.Sort(sortedkeys)
	for _, e := range sortedkeys {
		users = append(users, timekey[e])
	}

	return users
}

// FindTag takes a user's tweets and looks for a given tag.
// Returns the tweets with the tag as a []string.
func (userdata *Data) FindTag(tag string) {
	//for _, e := range userdata.status {
	//parts := strings.Split(e, "\t")

	//}
}
