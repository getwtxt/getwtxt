/*
Copyright (c) 2019 Ben Morrison (gbmor)

This file is part of Registry.

Registry is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

Registry is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with Registry.  If not, see <https://www.gnu.org/licenses/>.
*/

// Package registry implements functions and types that assist
// in the creation and management of a twtxt registry.
package registry // import "git.sr.ht/~gbmor/getwtxt/registry"

import (
	"net"
	"net/http"
	"sync"
	"time"
)

// Registrar implements the minimum amount of methods
// for a functioning Registry.
type Registrar interface {
	Put(user *User) error
	Get(urlKey string) (*User, error)
	DelUser(urlKey string) error
	UpdateUser(urlKey string) error
	GetUserStatuses(urlKey string) (TimeMap, error)
	GetStatuses() (TimeMap, error)
}

// User holds a given user's information
// and statuses.
type User struct {
	// Provided to aid in concurrency-safe
	// reads and writes. In most cases, the
	// mutex in the associated Index should be
	// used instead. This mutex is provided
	// should the library user need to access
	// a User independently of an Index.
	Mu sync.RWMutex

	// Nick is the user-specified nickname.
	Nick string

	// The URL of the user's twtxt file
	URL string

	// The reported last modification date
	// of the user's twtxt.txt file.
	LastModified string

	// The IP address of the user is optionally
	// recorded when submitted via POST.
	IP net.IP

	// The timestamp, in RFC3339 format,
	// reflecting when the user was added.
	Date string

	// A TimeMap of the user's statuses
	// from their twtxt file.
	Status TimeMap
}

// Registry enables the bulk of a registry's
// user data storage and access.
type Registry struct {
	// Provided to aid in concurrency-safe
	// reads and writes to a given registry
	// Users map.
	Mu sync.RWMutex

	// The registry's user data is contained
	// in this map. The functions within this
	// library expect the key to be the URL of
	// a given user's twtxt file.
	Users map[string]*User

	// The client to use for HTTP requests.
	// If nil is passed to NewIndex(), a
	// client with a 10 second timeout
	// and all other values as default is
	// used.
	HTTPClient *http.Client
}

// TimeMap holds extracted and processed user data as a
// string. A time.Time value is used as the key.
type TimeMap map[time.Time]string

// TimeSlice is a slice of time.Time used for sorting
// a TimeMap by timestamp.
type TimeSlice []time.Time

// NewUser returns a pointer to an initialized User
func NewUser() *User {
	return &User{
		Mu:     sync.RWMutex{},
		Status: NewTimeMap(),
	}
}

// New returns an initialized Registry instance.
func New(client *http.Client) *Registry {
	return &Registry{
		Mu:         sync.RWMutex{},
		Users:      make(map[string]*User),
		HTTPClient: client,
	}
}

// NewTimeMap returns an initialized TimeMap.
func NewTimeMap() TimeMap {
	return make(TimeMap)
}

// Len returns the length of the TimeSlice to be sorted.
// This helps satisfy sort.Interface.
func (t TimeSlice) Len() int {
	return len(t)
}

// Less returns true if the timestamp at index i is after
// the timestamp at index j in a given TimeSlice. This results
// in a descending (reversed) sort order for timestamps rather
// than ascending.
// This helps satisfy sort.Interface.
func (t TimeSlice) Less(i, j int) bool {
	return t[i].After(t[j])
}

// Swap transposes the timestamps at the two given indices
// for the TimeSlice receiver.
// This helps satisfy sort.Interface.
func (t TimeSlice) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}
