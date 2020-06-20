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
	"bufio"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"testing"
)

var addUserCases = []struct {
	name      string
	nick      string
	url       string
	wantErr   bool
	localOnly bool
}{
	{
		name:      "Legitimate User (Local Only)",
		nick:      "testuser1",
		url:       "http://localhost:8080/twtxt.txt",
		wantErr:   false,
		localOnly: true,
	},
	{
		name:      "Empty Query",
		nick:      "",
		url:       "",
		wantErr:   true,
		localOnly: false,
	},
	{
		name:      "Invalid URL",
		nick:      "foo",
		url:       "foobarringtons",
		wantErr:   true,
		localOnly: false,
	},
	{
		name:      "Garbage Data",
		nick:      "",
		url:       "",
		wantErr:   true,
		localOnly: false,
	},
}

// Tests if we can successfully add a user to the registry
func Test_Registry_AddUser(t *testing.T) {
	registry := initTestEnv()
	if !addUserCases[0].localOnly {
		http.Handle("/twtxt.txt", http.HandlerFunc(twtxtHandler))
		go fmt.Println(http.ListenAndServe(":8080", nil))
	}
	var buf = make([]byte, 256)
	// read random data into case 5
	rando, _ := os.Open("/dev/random")
	reader := bufio.NewReader(rando)
	n, err := reader.Read(buf)
	if err != nil || n == 0 {
		t.Errorf("Couldn't set up test: %v\n", err)
	}
	addUserCases[3].nick = string(buf)
	addUserCases[3].url = string(buf)

	statuses, err := registry.GetStatuses()
	if err != nil {
		t.Errorf("Error setting up test: %v\n", err)
	}

	for n, tt := range addUserCases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.localOnly {
				t.Skipf("Local-only test. Skipping ... ")
			}

			err := registry.AddUser(tt.nick, tt.url, nil, statuses)

			// only run some checks if we don't want an error
			if !tt.wantErr {
				if err != nil {
					t.Errorf("Got error: %v\n", err)
				}

				// make sure we have *something* in the registry
				if reflect.ValueOf(registry.Users[tt.url]).IsNil() {
					t.Errorf("Failed to add user %v registry.\n", tt.url)
				}

				// see if the nick in the registry is the same
				// as the test case. verifies the URL and the nick
				// since the URL is used as the key
				data := registry.Users[tt.url]
				if data.Nick != tt.nick {
					t.Errorf("Incorrect user data added to registry for user %v.\n", tt.url)
				}
			}
			// check for the cases that should throw an error
			if tt.wantErr && err == nil {
				t.Errorf("Expected error for case %v, got nil\n", n)
			}
		})
	}
}
func Benchmark_Registry_AddUser(b *testing.B) {
	registry := initTestEnv()
	statuses, err := registry.GetStatuses()
	if err != nil {
		b.Errorf("Error setting up test: %v\n", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tt := range addUserCases {
			err := registry.AddUser(tt.nick, tt.url, nil, statuses)
			if err != nil {
				continue
			}
			registry.Users[tt.url] = &User{}
		}
	}
}

var delUserCases = []struct {
	name    string
	url     string
	wantErr bool
}{
	{
		name:    "Valid User",
		url:     "https://example.com/twtxt.txt",
		wantErr: false,
	},
	{
		name:    "Valid User",
		url:     "https://example3.com/twtxt.txt",
		wantErr: false,
	},
	{
		name:    "Already Deleted User",
		url:     "https://example3.com/twtxt.txt",
		wantErr: true,
	},
	{
		name:    "Empty Query",
		url:     "",
		wantErr: true,
	},
	{
		name:    "Garbage Data",
		url:     "",
		wantErr: true,
	},
}

// Tests if we can successfully delete a user from the registry
func Test_Registry_DelUser(t *testing.T) {
	registry := initTestEnv()
	var buf = make([]byte, 256)
	// read random data into case 5
	rando, _ := os.Open("/dev/random")
	reader := bufio.NewReader(rando)
	n, err := reader.Read(buf)
	if err != nil || n == 0 {
		t.Errorf("Couldn't set up test: %v\n", err)
	}
	delUserCases[4].url = string(buf)

	for n, tt := range delUserCases {
		t.Run(tt.name, func(t *testing.T) {

			err := registry.DelUser(tt.url)
			if !reflect.ValueOf(registry.Users[tt.url]).IsNil() {
				t.Errorf("Failed to delete user %v from registry.\n", tt.url)
			}
			if tt.wantErr && err == nil {
				t.Errorf("Expected error but did not receive. Case %v\n", n)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error for case %v: %v\n", n, err)
			}
		})
	}
}
func Benchmark_Registry_DelUser(b *testing.B) {
	registry := initTestEnv()

	data1 := &User{
		Nick:   registry.Users[delUserCases[0].url].Nick,
		Date:   registry.Users[delUserCases[0].url].Date,
		Status: registry.Users[delUserCases[0].url].Status,
	}

	data2 := &User{
		Nick:   registry.Users[delUserCases[1].url].Nick,
		Date:   registry.Users[delUserCases[1].url].Date,
		Status: registry.Users[delUserCases[1].url].Status,
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, tt := range delUserCases {
			err := registry.DelUser(tt.url)
			if err != nil {
				continue
			}
		}

		registry.Users[delUserCases[0].url] = data1
		registry.Users[delUserCases[1].url] = data2
	}
}

var getUserStatusCases = []struct {
	name    string
	url     string
	wantErr bool
}{
	{
		name:    "Valid User",
		url:     "https://example.com/twtxt.txt",
		wantErr: false,
	},
	{
		name:    "Valid User",
		url:     "https://example3.com/twtxt.txt",
		wantErr: false,
	},
	{
		name:    "Nonexistent User",
		url:     "https://doesn't.exist/twtxt.txt",
		wantErr: true,
	},
	{
		name:    "Empty Query",
		url:     "",
		wantErr: true,
	},
	{
		name:    "Garbage Data",
		url:     "",
		wantErr: true,
	},
}

// Checks if we can retrieve a single user's statuses
func Test_Registry_GetUserStatuses(t *testing.T) {
	registry := initTestEnv()
	var buf = make([]byte, 256)
	// read random data into case 5
	rando, _ := os.Open("/dev/random")
	reader := bufio.NewReader(rando)
	n, err := reader.Read(buf)
	if err != nil || n == 0 {
		t.Errorf("Couldn't set up test: %v\n", err)
	}
	getUserStatusCases[4].url = string(buf)

	for n, tt := range getUserStatusCases {
		t.Run(tt.name, func(t *testing.T) {

			statuses, err := registry.GetUserStatuses(tt.url)

			if !tt.wantErr {
				if reflect.ValueOf(statuses).IsNil() {
					t.Errorf("Failed to pull statuses for user %v\n", tt.url)
				}
				// see if the function returns the same data
				// that we already have
				data := registry.Users[tt.url]
				if !reflect.DeepEqual(data.Status, statuses) {
					t.Errorf("Incorrect data retrieved as statuses for user %v.\n", tt.url)
				}
			}

			if tt.wantErr && err == nil {
				t.Errorf("Expected error, received nil for case %v: %v\n", n, tt.url)
			}
		})
	}
}
func Benchmark_Registry_GetUserStatuses(b *testing.B) {
	registry := initTestEnv()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, tt := range getUserStatusCases {
			_, err := registry.GetUserStatuses(tt.url)
			if err != nil {
				continue
			}
		}
	}
}

// Tests if we can retrieve all user statuses at once
func Test_Registry_GetStatuses(t *testing.T) {
	registry := initTestEnv()
	t.Run("Registry.GetStatuses()", func(t *testing.T) {

		statuses, err := registry.GetStatuses()
		if reflect.ValueOf(statuses).IsNil() || err != nil {
			t.Errorf("Failed to pull all statuses. %v\n", err)
		}

		// Now do the same query manually to see
		// if we get the same result
		unionmap := NewTimeMap()
		for _, v := range registry.Users {
			for i, e := range v.Status {
				unionmap[i] = e
			}
		}
		if !reflect.DeepEqual(statuses, unionmap) {
			t.Errorf("Incorrect data retrieved as statuses.\n")
		}
	})
}
func Benchmark_Registry_GetStatuses(b *testing.B) {
	registry := initTestEnv()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := registry.GetStatuses()
		if err != nil {
			continue
		}
	}
}
