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
	"os"
	"strings"
	"testing"
	"time"
)

var queryUserCases = []struct {
	name    string
	term    string
	wantErr bool
}{
	{
		name:    "Valid User",
		term:    "foo",
		wantErr: false,
	},
	{
		name:    "Empty Query",
		term:    "",
		wantErr: false,
	},
	{
		name:    "Nonexistent User",
		term:    "doesntexist",
		wantErr: true,
	},
	{
		name:    "Garbage Data",
		term:    "will be replaced with garbage data",
		wantErr: true,
	},
}

// Checks if Registry.QueryUser() returns users that
// match the provided substring.
func Test_Registry_QueryUser(t *testing.T) {
	registry := initTestEnv()
	var buf = make([]byte, 256)
	// read random data into case 8
	rando, _ := os.Open("/dev/random")
	reader := bufio.NewReader(rando)
	n, err := reader.Read(buf)
	if err != nil || n == 0 {
		t.Errorf("Couldn't set up test: %v\n", err)
	}
	queryUserCases[3].term = string(buf)

	for n, tt := range queryUserCases {

		t.Run(tt.name, func(t *testing.T) {
			out, err := registry.QueryUser(tt.term)

			if out == nil && err != nil && !tt.wantErr {
				t.Errorf("Received nil output or an error when unexpected. Case %v, %v, %v\n", n, tt.term, err)
			}

			if out != nil && tt.wantErr {
				t.Errorf("Received unexpected nil output when an error was expected. Case %v, %v\n", n, tt.term)
			}

			for _, e := range out {
				one := strings.Split(e, "\t")

				if !strings.Contains(one[0], tt.term) && !strings.Contains(one[1], tt.term) {
					t.Errorf("Received incorrect output: %v != %v\n", tt.term, e)
				}
			}
		})
	}
}
func Benchmark_Registry_QueryUser(b *testing.B) {
	registry := initTestEnv()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, tt := range queryUserCases {
			_, err := registry.QueryUser(tt.term)
			if err != nil {
				b.Errorf("%v\n", err)
			}
		}
	}
}

var queryInStatusCases = []struct {
	name    string
	substr  string
	wantNil bool
	wantErr bool
}{
	{
		name:    "Tag in Status",
		substr:  "twtxt",
		wantNil: false,
		wantErr: false,
	},
	{
		name:    "Valid URL",
		substr:  "https://example.com/twtxt.txt",
		wantNil: false,
		wantErr: false,
	},
	{
		name:    "Multiple Words in Status",
		substr:  "next programming",
		wantNil: false,
		wantErr: false,
	},
	{
		name:    "Multiple Words, Not in Status",
		substr:  "explosive bananas from antarctica",
		wantNil: true,
		wantErr: false,
	},
	{
		name:    "Empty Query",
		substr:  "",
		wantNil: true,
		wantErr: true,
	},
	{
		name:    "Nonsense",
		substr:  "ahfiurrenkhfkajdhfao",
		wantNil: true,
		wantErr: false,
	},
	{
		name:    "Invalid URL",
		substr:  "https://doesnt.exist/twtxt.txt",
		wantNil: true,
		wantErr: false,
	},
	{
		name:    "Garbage Data",
		substr:  "will be replaced with garbage data",
		wantNil: true,
		wantErr: false,
	},
}

// This tests whether we can find a substring in all of
// the known status messages, disregarding the metadata
// stored with each status.
func Test_Registry_QueryInStatus(t *testing.T) {
	registry := initTestEnv()
	var buf = make([]byte, 256)
	// read random data into case 8
	rando, _ := os.Open("/dev/random")
	reader := bufio.NewReader(rando)
	n, err := reader.Read(buf)
	if err != nil || n == 0 {
		t.Errorf("Couldn't set up test: %v\n", err)
	}
	queryInStatusCases[7].substr = string(buf)

	for _, tt := range queryInStatusCases {

		t.Run(tt.name, func(t *testing.T) {

			out, err := registry.QueryInStatus(tt.substr)
			if err != nil && !tt.wantErr {
				t.Errorf("Caught unexpected error: %v\n", err)
			}

			if !tt.wantErr && out == nil && !tt.wantNil {
				t.Errorf("Got nil when expecting output\n")
			}

			if err == nil && tt.wantErr {
				t.Errorf("Expecting error, got nil.\n")
			}

			for _, e := range out {
				split := strings.Split(strings.ToLower(e), "\t")

				if e != "" {
					if !strings.Contains(split[3], strings.ToLower(tt.substr)) {
						t.Errorf("Status without substring returned\n")
					}
				}
			}
		})
	}

}
func Benchmark_Registry_QueryInStatus(b *testing.B) {
	registry := initTestEnv()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, tt := range queryInStatusCases {
			_, err := registry.QueryInStatus(tt.substr)
			if err != nil {
				continue
			}
		}
	}
}

// Tests whether we can retrieve the 20 most
// recent statuses in the registry
func Test_QueryAllStatuses(t *testing.T) {
	registry := initTestEnv()
	t.Run("Latest Statuses", func(t *testing.T) {
		out, err := registry.QueryAllStatuses()
		if out == nil || err != nil {
			t.Errorf("Got no statuses, or more than 20: %v, %v\n", len(out), err)
		}
	})
}
func Benchmark_QueryAllStatuses(b *testing.B) {
	registry := initTestEnv()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := registry.QueryAllStatuses()
		if err != nil {
			continue
		}
	}
}

var get20cases = []struct {
	name    string
	page    int
	wantErr bool
}{
	{
		name:    "First Page",
		page:    1,
		wantErr: false,
	},
	{
		name:    "High Page Number",
		page:    256,
		wantErr: false,
	},
	{
		name:    "Illegal Page Number",
		page:    -23,
		wantErr: false,
	},
}

func Test_ReduceToPage(t *testing.T) {
	registry := initTestEnv()
	for _, tt := range get20cases {
		t.Run(tt.name, func(t *testing.T) {
			out, err := registry.QueryAllStatuses()
			if err != nil && !tt.wantErr {
				t.Errorf("%v\n", err.Error())
			}
			out = ReduceToPage(tt.page, out)
			if len(out) > 20 || len(out) == 0 {
				t.Errorf("Page-Reduce Malfunction: length of data %v\n", len(out))
			}
		})
	}
}

func Benchmark_ReduceToPage(b *testing.B) {
	registry := initTestEnv()
	out, _ := registry.QueryAllStatuses()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tt := range get20cases {
			ReduceToPage(tt.page, out)
		}
	}
}

// This tests whether we can find a substring in the
// given user's status messages, disregarding the metadata
// stored with each status.
func Test_User_FindInStatus(t *testing.T) {
	registry := initTestEnv()
	var buf = make([]byte, 256)
	// read random data into case 8
	rando, _ := os.Open("/dev/random")
	reader := bufio.NewReader(rando)
	n, err := reader.Read(buf)
	if err != nil || n == 0 {
		t.Errorf("Couldn't set up test: %v\n", err)
	}
	queryInStatusCases[7].substr = string(buf)

	data := make([]*User, 0)

	for _, v := range registry.Users {
		data = append(data, v)
	}

	for _, tt := range queryInStatusCases {
		t.Run(tt.name, func(t *testing.T) {
			for _, e := range data {

				tag := e.FindInStatus(tt.substr)
				if tag == nil && !tt.wantNil {
					t.Errorf("Got nil tag\n")
				}
			}
		})
	}

}
func Benchmark_User_FindInStatus(b *testing.B) {
	registry := initTestEnv()
	data := make([]*User, 0)

	for _, v := range registry.Users {
		data = append(data, v)
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, tt := range data {
			for _, v := range queryInStatusCases {
				tt.FindInStatus(v.substr)
			}
		}
	}
}

func Test_SortByTime_Slice(t *testing.T) {
	registry := initTestEnv()

	statusmap, err := registry.GetStatuses()
	if err != nil {
		t.Errorf("Failed to finish test initialization: %v\n", err)
	}

	t.Run("Sort By Time ([]TimeMap)", func(t *testing.T) {
		sorted, err := SortByTime(statusmap)
		if err != nil {
			t.Errorf("%v\n", err)
		}
		split := strings.Split(sorted[0], "\t")
		firsttime, _ := time.Parse("RFC3339", split[0])

		for i := range sorted {
			if i < len(sorted)-1 {

				nextsplit := strings.Split(sorted[i+1], "\t")
				nexttime, _ := time.Parse("RFC3339", nextsplit[0])

				if firsttime.Before(nexttime) {
					t.Errorf("Timestamps out of order: %v\n", sorted)
				}

				firsttime = nexttime
			}
		}
	})
}

// Benchmarking a sort of 1000000 statuses by timestamp.
// Right now it's at roughly 2000ns per 2 statuses.
// Set sortMultiplier to be the number of desired
// statuses divided by four.
func Benchmark_SortByTime_Slice(b *testing.B) {
	// I set this to 250,000,000 and it hard-locked
	// my laptop. Oops.
	sortMultiplier := 250
	b.Logf("Benchmarking SortByTime with a constructed slice of %v statuses ...\n", sortMultiplier*4)
	registry := initTestEnv()

	statusmap, err := registry.GetStatuses()
	if err != nil {
		b.Errorf("Failed to finish benchmark initialization: %v\n", err)
	}

	// Constructed registry has four statuses. This
	// makes a TimeMapSlice of 1000000 statuses.
	statusmaps := make([]TimeMap, sortMultiplier*4)
	for i := 0; i < sortMultiplier; i++ {
		statusmaps = append(statusmaps, statusmap)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := SortByTime(statusmaps...)
		if err != nil {
			b.Errorf("%v\n", err)
		}
	}
}

func Test_SortByTime_Single(t *testing.T) {
	registry := initTestEnv()

	statusmap, err := registry.GetStatuses()
	if err != nil {
		t.Errorf("Failed to finish test initialization: %v\n", err)
	}

	t.Run("Sort By Time (TimeMap)", func(t *testing.T) {
		sorted, err := SortByTime(statusmap)
		if err != nil {
			t.Errorf("%v\n", err)
		}
		split := strings.Split(sorted[0], "\t")
		firsttime, _ := time.Parse("RFC3339", split[0])

		for i := range sorted {
			if i < len(sorted)-1 {

				nextsplit := strings.Split(sorted[i+1], "\t")
				nexttime, _ := time.Parse("RFC3339", nextsplit[0])

				if firsttime.Before(nexttime) {
					t.Errorf("Timestamps out of order: %v\n", sorted)
				}

				firsttime = nexttime
			}
		}
	})
}

func Benchmark_SortByTime_Single(b *testing.B) {
	registry := initTestEnv()

	statusmap, err := registry.GetStatuses()
	if err != nil {
		b.Errorf("Failed to finish benchmark initialization: %v\n", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := SortByTime(statusmap)
		if err != nil {
			b.Errorf("%v\n", err)
		}
	}
}
