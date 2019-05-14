package registry

import (
	"log"
	"time"
)

// AddUser inserts a new user into the index. The *Data struct
// contains the nickname and the time the user was added.
func (index UserIndex) AddUser(nick string, url string) {
	rfc3339date, err := time.Now().MarshalText()
	if err != nil {
		log.Printf("Error formatting user add time as RFC3339: %v\n", err)
	}
	imutex.Lock()
	index[url] = &Data{Nick: nick, Date: time.Now(), APIdate: rfc3339date}
	imutex.Unlock()
}

// DelUser removes a user from the index completely.
func (index UserIndex) DelUser(url string) {
	imutex.Lock()
	delete(index, url)
	imutex.Unlock()
}

// GetUserStatuses returns a TimeMap containing a user's statuses
func (index UserIndex) GetUserStatuses(url string) TimeMap {
	imutex.RLock()
	status := index[url].Status
	imutex.RUnlock()
	return status
}

// GetStatuses returns a TimeMap containing all statuses
func (index UserIndex) GetStatuses() TimeMap {
	statuses := NewTimeMap()
	imutex.RLock()
	for _, v := range index {
		for a, b := range v.Status {
			statuses[a] = b
		}
	}
	imutex.RUnlock()
	return statuses
}
