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
	"sort"
	"strings"
	"time"
)

// QueryUser checks the Registry for usernames
// or user URLs that contain the term provided as an argument. Entries
// are returned sorted by the date they were added to the Registry. If
// the argument provided is blank, return all users.
func (registry *Registry) QueryUser(term string) ([]string, error) {
	if registry == nil {
		return nil, fmt.Errorf("can't query empty registry for user")
	}

	term = strings.ToLower(term)
	timekey := NewTimeMap()
	keys := make(TimeSlice, 0)
	var users []string

	registry.Mu.RLock()
	defer registry.Mu.RUnlock()

	for k, v := range registry.Users {
		if registry.Users[k] == nil {
			continue
		}
		v.Mu.RLock()
		if strings.Contains(strings.ToLower(v.Nick), term) || strings.Contains(strings.ToLower(k), term) {
			thetime, err := time.Parse(time.RFC3339, v.Date)
			if err != nil {
				v.Mu.RUnlock()
				continue
			}
			timekey[thetime] = v.Nick + "\t" + k + "\t" + v.Date + "\n"
			keys = append(keys, thetime)
		}
		v.Mu.RUnlock()
	}

	sort.Sort(keys)
	for _, e := range keys {
		users = append(users, timekey[e])
	}

	return users, nil
}

// QueryInStatus returns all statuses in the Registry
// that contain the provided substring (tag, mention URL, etc).
func (registry *Registry) QueryInStatus(substring string) ([]string, error) {
	if substring == "" {
		return nil, fmt.Errorf("cannot query for empty tag")
	} else if registry == nil {
		return nil, fmt.Errorf("can't query statuses of empty registry")
	}

	statusmap := make([]TimeMap, 0)

	registry.Mu.RLock()
	defer registry.Mu.RUnlock()

	for _, v := range registry.Users {
		statusmap = append(statusmap, v.FindInStatus(substring))
	}

	sorted, err := SortByTime(statusmap...)
	if err != nil {
		return nil, err
	}

	return sorted, nil
}

// QueryAllStatuses returns all statuses in the Registry
// as a slice of strings sorted by timestamp.
func (registry *Registry) QueryAllStatuses() ([]string, error) {
	if registry == nil {
		return nil, fmt.Errorf("can't get latest statuses from empty registry")
	}

	statusmap, err := registry.GetStatuses()
	if err != nil {
		return nil, err
	}

	sorted, err := SortByTime(statusmap)
	if err != nil {
		return nil, err
	}

	if sorted == nil {
		sorted = make([]string, 1)
	}

	return sorted, nil
}

// ReduceToPage returns the passed 'page' worth of output.
// One page is twenty items. For example, if 2 is passed,
// it will return data[20:40]. According to the twtxt
// registry specification, queries should accept a "page"
// value.
func ReduceToPage(page int, data []string) []string {
	end := 20 * page
	if end > len(data) || end < 1 {
		end = len(data)
	}

	beg := end - 20
	if beg > len(data)-1 || beg < 0 {
		beg = 0
	}

	return data[beg:end]
}

// FindInStatus takes a user's statuses and looks for a given substring.
// Returns the statuses that include the substring as a TimeMap.
func (userdata *User) FindInStatus(substring string) TimeMap {
	if userdata == nil {
		return nil
	} else if len(substring) > 140 {
		return nil
	}

	substring = strings.ToLower(substring)
	statuses := NewTimeMap()

	userdata.Mu.RLock()
	defer userdata.Mu.RUnlock()

	for k, e := range userdata.Status {
		if _, ok := userdata.Status[k]; !ok {
			continue
		}

		parts := strings.Split(strings.ToLower(e), "\t")
		if strings.Contains(parts[3], substring) {
			statuses[k] = e
		}
	}

	return statuses
}

// SortByTime returns a string slice of the query results,
// sorted by timestamp in descending order (newest first).
func SortByTime(tm ...TimeMap) ([]string, error) {
	if tm == nil {
		return nil, fmt.Errorf("can't sort nil TimeMaps")
	}

	var times = make(TimeSlice, 0)
	var data []string

	for _, e := range tm {
		for k := range e {
			times = append(times, k)
		}
	}

	sort.Sort(times)

	for k := range tm {
		for _, e := range times {
			if _, ok := tm[k][e]; ok {
				data = append(data, tm[k][e])
			}
		}
	}

	return data, nil
}
