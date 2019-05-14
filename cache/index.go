package cache

import (
	"log"
	"strings"
	"time"
)

// NewUserIndex returns a new instance of a user index
func NewUserIndex() *UserIndex {
	return &UserIndex{}
}

// AddUser inserts a new user into the index. The *Data struct only contains the nickname.)
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

func (index UserIndex) QueryUser(name string) []string {
	var users []string
	var entry string
	for k, v := range index {
		if strings.Contains(v.nick, name) {
			entry = v.nick + "\t" + k + "\t" + string(v.apidate)
			users = append(users, entry)
		}
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
