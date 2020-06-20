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

package registry // import "git.sr.ht/~gbmor/getwtxt/registry"

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

// AddUser inserts a new user into the Registry.
func (registry *Registry) AddUser(nickname, urlKey string, ipAddress net.IP, statuses TimeMap) error {

	if registry == nil {
		return fmt.Errorf("can't add user to uninitialized registry")

	} else if nickname == "" || urlKey == "" {
		return fmt.Errorf("both URL and Nick must be specified")

	} else if !strings.HasPrefix(urlKey, "http") {
		return fmt.Errorf("invalid URL: %v", urlKey)
	}

	registry.Mu.Lock()
	defer registry.Mu.Unlock()

	if _, ok := registry.Users[urlKey]; ok {
		return fmt.Errorf("user %v already exists", urlKey)
	}

	registry.Users[urlKey] = &User{
		Mu:           sync.RWMutex{},
		Nick:         nickname,
		URL:          urlKey,
		LastModified: "",
		IP:           ipAddress,
		Date:         time.Now().Format(time.RFC3339),
		Status:       statuses}

	return nil
}

// Put inserts a given User into an Registry. The User
// being pushed need only have the URL field filled.
// All other fields may be empty.
// This can be destructive: an existing User in the
// Registry will be overwritten if its User.URL is the
// same as the User.URL being pushed.
func (registry *Registry) Put(user *User) error {
	if user == nil {
		return fmt.Errorf("can't push nil data to registry")
	}
	if registry == nil || registry.Users == nil {
		return fmt.Errorf("can't push data to registry: registry uninitialized")
	}
	user.Mu.RLock()
	if user.URL == "" {
		user.Mu.RUnlock()
		return fmt.Errorf("can't push data to registry: missing URL for key")
	}
	urlKey := user.URL
	registry.Mu.Lock()
	registry.Users[urlKey] = user
	registry.Mu.Unlock()
	user.Mu.RUnlock()

	return nil
}

// Get returns the User associated with the
// provided URL key in the Registry.
func (registry *Registry) Get(urlKey string) (*User, error) {
	if registry == nil {
		return nil, fmt.Errorf("can't pop from nil registry")
	}
	if urlKey == "" {
		return nil, fmt.Errorf("can't pop unless provided a key")
	}

	registry.Mu.RLock()
	defer registry.Mu.RUnlock()

	if _, ok := registry.Users[urlKey]; !ok {
		return nil, fmt.Errorf("provided url key doesn't exist in registry")
	}

	registry.Users[urlKey].Mu.RLock()
	userGot := registry.Users[urlKey]
	registry.Users[urlKey].Mu.RUnlock()

	return userGot, nil
}

// DelUser removes a user and all associated data from
// the Registry.
func (registry *Registry) DelUser(urlKey string) error {

	if registry == nil {
		return fmt.Errorf("can't delete user from empty registry")

	} else if urlKey == "" {
		return fmt.Errorf("can't delete blank user")

	} else if !strings.HasPrefix(urlKey, "http") {
		return fmt.Errorf("invalid URL: %v", urlKey)
	}

	registry.Mu.Lock()
	defer registry.Mu.Unlock()

	if _, ok := registry.Users[urlKey]; !ok {
		return fmt.Errorf("can't delete user %v, user doesn't exist", urlKey)
	}

	delete(registry.Users, urlKey)

	return nil
}

// UpdateUser scrapes an existing user's remote twtxt.txt
// file. Any new statuses are added to the user's entry
// in the Registry. If the remote twtxt data's reported
// Content-Length does not differ from what is stored,
// an error is returned.
func (registry *Registry) UpdateUser(urlKey string) error {
	if urlKey == "" || !strings.HasPrefix(urlKey, "http") {
		return fmt.Errorf("invalid URL: %v", urlKey)
	}

	diff, err := registry.DiffTwtxt(urlKey)
	if err != nil {
		return err
	} else if !diff {
		return fmt.Errorf("no new statuses available for %v", urlKey)
	}

	out, isRemoteRegistry, err := GetTwtxt(urlKey, registry.HTTPClient)
	if err != nil {
		return err
	}

	if isRemoteRegistry {
		return fmt.Errorf("attempting to update registry URL - users should be updated individually")
	}

	registry.Mu.Lock()
	defer registry.Mu.Unlock()
	user := registry.Users[urlKey]

	user.Mu.Lock()
	defer user.Mu.Unlock()
	nick := user.Nick

	data, err := ParseUserTwtxt(out, nick, urlKey)
	if err != nil {
		return err
	}

	for i, e := range data {
		user.Status[i] = e
	}

	registry.Users[urlKey] = user

	return nil
}

// CrawlRemoteRegistry scrapes all nicknames and user URLs
// from a provided registry. The urlKey passed to this function
// must be in the form of https://registry.example.com/api/plain/users
func (registry *Registry) CrawlRemoteRegistry(urlKey string) error {
	if urlKey == "" || !strings.HasPrefix(urlKey, "http") {
		return fmt.Errorf("invalid URL: %v", urlKey)
	}

	out, isRemoteRegistry, err := GetTwtxt(urlKey, registry.HTTPClient)
	if err != nil {
		return err
	}

	if !isRemoteRegistry {
		return fmt.Errorf("can't add single user via call to CrawlRemoteRegistry")
	}

	users, err := ParseRegistryTwtxt(out)
	if err != nil {
		return err
	}

	// only add new users so we don't overwrite data
	// we already have (and lose statuses, etc)
	registry.Mu.Lock()
	defer registry.Mu.Unlock()
	for _, e := range users {
		if _, ok := registry.Users[e.URL]; !ok {
			registry.Users[e.URL] = e
		}
	}

	return nil
}

// GetUserStatuses returns a TimeMap containing single user's statuses
func (registry *Registry) GetUserStatuses(urlKey string) (TimeMap, error) {
	if registry == nil {
		return nil, fmt.Errorf("can't get statuses from an empty registry")
	} else if urlKey == "" || !strings.HasPrefix(urlKey, "http") {
		return nil, fmt.Errorf("invalid URL: %v", urlKey)
	}

	registry.Mu.RLock()
	defer registry.Mu.RUnlock()
	if _, ok := registry.Users[urlKey]; !ok {
		return nil, fmt.Errorf("can't retrieve statuses of nonexistent user")
	}

	registry.Users[urlKey].Mu.RLock()
	status := registry.Users[urlKey].Status
	registry.Users[urlKey].Mu.RUnlock()

	return status, nil
}

// GetStatuses returns a TimeMap containing all statuses
// from all users in the Registry.
func (registry *Registry) GetStatuses() (TimeMap, error) {
	if registry == nil {
		return nil, fmt.Errorf("can't get statuses from an empty registry")
	}

	statuses := NewTimeMap()

	registry.Mu.RLock()
	defer registry.Mu.RUnlock()

	for _, v := range registry.Users {
		v.Mu.RLock()
		if v.Status == nil || len(v.Status) == 0 {
			v.Mu.RUnlock()
			continue
		}
		for a, b := range v.Status {
			if _, ok := v.Status[a]; ok {
				statuses[a] = b
			}
		}
		v.Mu.RUnlock()
	}

	return statuses, nil
}
