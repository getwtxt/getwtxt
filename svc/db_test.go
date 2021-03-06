/*
Copyright (c) 2019 Ben Morrison (gbmor)

This file is part of Getwtxt.

Getwtxt is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

Getwtxt is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with Getwtxt.  If not, see <https://www.gnu.org/licenses/>.
*/

package svc // import "git.sr.ht/~gbmor/getwtxt/svc"

import (
	"net"
	"testing"

	"git.sr.ht/~gbmor/getwtxt/registry"
)

func Test_pushpullDatabase(t *testing.T) {
	initTestConf()
	initTestDB()

	out, _, err := registry.GetTwtxt(testTwtxtURL, nil)
	if err != nil {
		t.Errorf("Couldn't set up test: %v\n", err)
	}

	statusmap, err := registry.ParseUserTwtxt(out, "getwtxttest", testTwtxtURL)
	if err != nil {
		t.Errorf("Couldn't set up test: %v\n", err)
	}

	twtxtCache.AddUser("getwtxttest", testTwtxtURL, net.ParseIP("127.0.0.1"), statusmap)

	remoteRegistries.List = append(remoteRegistries.List, "https://twtxt.tilde.institute/api/plain/users")

	t.Run("Push to Database", func(t *testing.T) {
		err := pushDB()
		if err != nil {
			t.Errorf("%v\n", err)
		}
	})

	t.Run("Clearing Registry", func(t *testing.T) {
		err := twtxtCache.DelUser(testTwtxtURL)
		if err != nil {
			t.Errorf("%v", err)
		}
	})

	t.Run("Pulling from Database", func(t *testing.T) {
		pullDB()

		twtxtCache.Mu.RLock()
		if _, ok := twtxtCache.Users[testTwtxtURL]; !ok {
			t.Errorf("Missing user previously pushed to database\n")
		}
		twtxtCache.Mu.RUnlock()

	})
}
func Benchmark_pushDatabase(b *testing.B) {
	initTestConf()
	initTestDB()

	if _, ok := twtxtCache.Users[testTwtxtURL]; !ok {
		out, _, err := registry.GetTwtxt(testTwtxtURL, nil)
		if err != nil {
			b.Errorf("Couldn't set up benchmark: %v\n", err)
		}

		statusmap, err := registry.ParseUserTwtxt(out, "getwtxttest", testTwtxtURL)
		if err != nil {
			b.Errorf("Couldn't set up benchmark: %v\n", err)
		}

		twtxtCache.AddUser("getwtxttest", testTwtxtURL, net.ParseIP("127.0.0.1"), statusmap)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := pushDB()
		if err != nil {
			b.Errorf("%v\n", err)
		}
	}
}
func Benchmark_pullDatabase(b *testing.B) {
	initTestConf()
	initTestDB()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		pullDB()
	}
}
