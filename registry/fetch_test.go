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

package registry

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

func constructTwtxt() []byte {
	registry := initTestEnv()
	var resp []byte
	// iterates through each mock user's mock statuses
	for _, v := range registry.Users {
		for _, e := range v.Status {
			split := strings.Split(e, "\t")
			status := []byte(split[2] + "\t" + split[3] + "\n")
			resp = append(resp, status...)
		}
	}
	return resp
}

// this is just dumping all the mock statuses.
// it'll be served under fake paths as
// "remote" twtxt.txt files
func twtxtHandler(w http.ResponseWriter, _ *http.Request) {
	// prepare the response
	resp := constructTwtxt()
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	n, err := w.Write(resp)
	if err != nil || n == 0 {
		fmt.Printf("Got error or wrote zero bytes: %v bytes, %v\n", n, err)
	}
}

var getTwtxtCases = []struct {
	name      string
	url       string
	wantErr   bool
	localOnly bool
}{
	{
		name:      "Constructed Local twtxt.txt",
		url:       "http://localhost:8080/twtxt.txt",
		wantErr:   false,
		localOnly: true,
	},
	{
		name:      "Inaccessible Site With twtxt.txt",
		url:       "https://example33333333333.com/twtxt.txt",
		wantErr:   true,
		localOnly: false,
	},
	{
		name:      "Inaccessible Site Without twtxt.txt",
		url:       "https://example333333333333.com",
		wantErr:   true,
		localOnly: false,
	},
	{
		name:      "Local File Inclusion 1",
		url:       "file://init_test.go",
		wantErr:   true,
		localOnly: false,
	},
	{
		name:      "Local File Inclusion 2",
		url:       "/etc/passwd",
		wantErr:   true,
		localOnly: false,
	},
	{
		name:      "Remote File Inclusion",
		url:       "https://example.com/file.cgi",
		wantErr:   true,
		localOnly: false,
	},
	{
		name:      "Remote Registry",
		url:       "https://twtxt.tilde.institute/api/plain/tweets/",
		wantErr:   false,
		localOnly: false,
	},
	{
		name:      "Garbage Data",
		url:       "this will be replaced with garbage data",
		wantErr:   true,
		localOnly: true,
	},
}

// Test the function that yoinks the /twtxt.txt file
// for a given user.
func Test_GetTwtxt(t *testing.T) {
	var buf = make([]byte, 256)
	// read random data into case 4
	rando, _ := os.Open("/dev/random")
	reader := bufio.NewReader(rando)
	n, err := reader.Read(buf)
	if err != nil || n == 0 {
		t.Errorf("Couldn't set up test: %v\n", err)
	}
	getTwtxtCases[7].url = string(buf)

	if !getTwtxtCases[0].localOnly {
		http.Handle("/twtxt.txt", http.HandlerFunc(twtxtHandler))
		go fmt.Println(http.ListenAndServe(":8080", nil))
	}

	for _, tt := range getTwtxtCases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.localOnly {
				t.Skipf("Local-only test. Skipping ... \n")
			}
			out, _, err := GetTwtxt(tt.url, nil)
			if tt.wantErr && err == nil {
				t.Errorf("Expected error: %v\n", tt.url)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v %v\n", tt.url, err)
			}
			if !tt.wantErr && out == nil {
				t.Errorf("Incorrect data received: %v\n", out)
			}
		})
	}

}

// running the benchmarks separately for each case
// as they have different properties (allocs, time)
func Benchmark_GetTwtxt(b *testing.B) {

	for i := 0; i < b.N; i++ {
		_, _, err := GetTwtxt("https://gbmor.dev/twtxt.txt", nil)
		if err != nil {
			continue
		}
	}
}

var parseTwtxtCases = []struct {
	name      string
	data      []byte
	wantErr   bool
	localOnly bool
}{
	{
		name:      "Constructed twtxt file",
		data:      constructTwtxt(),
		wantErr:   false,
		localOnly: false,
	},
	{
		name:      "Incorrectly formatted date",
		data:      []byte("2019 April 23rd\tI love twtxt!!!11"),
		wantErr:   true,
		localOnly: false,
	},
	{
		name:      "No data",
		data:      []byte{},
		wantErr:   true,
		localOnly: false,
	},
	{
		name:      "Variant rfc3339 datestamp",
		data:      []byte("2020-02-04T21:28:21.868659+00:00\tWill this work?"),
		wantErr:   false,
		localOnly: false,
	},
	{
		name:      "Random/garbage data",
		wantErr:   true,
		localOnly: true,
	},
}

// See if we can break ParseTwtxt or get it
// to throw an unexpected error
func Test_ParseUserTwtxt(t *testing.T) {
	var buf = make([]byte, 256)
	// read random data into case 4
	rando, _ := os.Open("/dev/random")
	reader := bufio.NewReader(rando)
	n, err := reader.Read(buf)
	if err != nil || n == 0 {
		t.Errorf("Couldn't set up test: %v\n", err)
	}
	parseTwtxtCases[4].data = buf

	for _, tt := range parseTwtxtCases {
		if tt.localOnly {
			t.Skipf("Local-only test: Skipping ... \n")
		}
		t.Run(tt.name, func(t *testing.T) {
			timemap, errs := ParseUserTwtxt(tt.data, "testuser", "testurl")
			if errs == nil && tt.wantErr {
				t.Errorf("Expected error(s), received none.\n")
			}

			if !tt.wantErr {
				if errs != nil {
					t.Errorf("Unexpected error: %v\n", errs)
				}

				for k, v := range timemap {
					if k == (time.Time{}) || v == "" {
						t.Errorf("Empty status or empty timestamp: %v, %v\n", k, v)
					}
				}
			}
		})
	}
}

func Benchmark_ParseUserTwtxt(b *testing.B) {
	var buf = make([]byte, 256)
	// read random data into case 4
	rando, _ := os.Open("/dev/random")
	reader := bufio.NewReader(rando)
	n, err := reader.Read(buf)
	if err != nil || n == 0 {
		b.Errorf("Couldn't set up benchmark: %v\n", err)
	}
	parseTwtxtCases[3].data = buf

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, tt := range parseTwtxtCases {
			_, _ = ParseUserTwtxt(tt.data, "testuser", "testurl")
		}
	}
}

var timestampCases = []struct {
	name     string
	orig     string
	expected string
}{
	{
		name:     "Timezone appended",
		orig:     "2020-01-13T16:08:25.544735+00:00",
		expected: "2020-01-13T16:08:25.544735Z",
	},
	{
		name:     "It's fine already",
		orig:     "2020-01-14T00:19:45.092344Z",
		expected: "2020-01-14T00:19:45.092344Z",
	},
}
