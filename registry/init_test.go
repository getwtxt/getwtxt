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

package registry //import "git.sr.ht/~gbmor/getwtxt/registry"

import (
	"fmt"
	"log"
	"os"
	"time"
)

func quickErr(err error) {
	if err != nil {
		fmt.Printf("%v\n", err)
	}
}

// Sets up mock users and statuses
func initTestEnv() *Registry {
	hush, err := os.Open("/dev/null")
	quickErr(err)
	log.SetOutput(hush)

	// this is a bit tedious, but set up fake dates
	// for the mock users' join and status timestamps
	timeMonthPrev := time.Now().AddDate(0, -1, 0)
	timeMonthPrevRFC := timeMonthPrev.Format(time.RFC3339)

	timeTwoMonthsPrev := time.Now().AddDate(0, -2, 0)
	timeTwoMonthsPrevRFC := timeTwoMonthsPrev.Format(time.RFC3339)

	timeThreeMonthsPrev := time.Now().AddDate(0, -3, 0)
	timeThreeMonthsPrevRFC := timeThreeMonthsPrev.Format(time.RFC3339)

	timeFourMonthsPrev := time.Now().AddDate(0, -4, 0)
	timeFourMonthsPrevRFC := timeFourMonthsPrev.Format(time.RFC3339)

	var mockusers = []struct {
		url     string
		nick    string
		date    string
		apidate []byte
		status  TimeMap
	}{
		{
			url:  "https://example3.com/twtxt.txt",
			nick: "foo_barrington",
			date: timeTwoMonthsPrevRFC,
			status: TimeMap{
				timeTwoMonthsPrev: "foo_barrington\thttps://example3.com/twtxt.txt\t" + timeTwoMonthsPrevRFC + "\tJust got started with #twtxt!",
				timeMonthPrev:     "foo_barrington\thttps://example3.com/twtxt.txt\t" + timeMonthPrevRFC + "\tHey <@foo https://example.com/twtxt.txt>, I love programming. Just FYI.",
			},
		},
		{
			url:  "https://example.com/twtxt.txt",
			nick: "foo",
			date: timeFourMonthsPrevRFC,
			status: TimeMap{
				timeFourMonthsPrev:  "foo\thttps://example.com/twtxt.txt\t" + timeFourMonthsPrevRFC + "\tThis is so much better than #twitter",
				timeThreeMonthsPrev: "foo\thttps://example.com/twtxt.txt\t" + timeThreeMonthsPrevRFC + "\tI can't wait to start on my next programming #project with <@foo_barrington https://example3.com/twtxt.txt>",
			},
		},
	}
	registry := New(nil)

	// fill the test registry with the mock users
	for _, e := range mockusers {
		data := &User{}
		data.Nick = e.nick
		data.Date = e.date
		data.Status = e.status
		registry.Users[e.url] = data
	}

	return registry
}
