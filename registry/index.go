package registry

import (
	"log"
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
