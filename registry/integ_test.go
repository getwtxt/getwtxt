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
	"strings"
	"testing"
)

// This tests all the operations on an registry.
func Test_Integration(t *testing.T) {
	var integration = func(t *testing.T) {
		t.Logf("Creating registry object ...\n")
		registry := New(nil)

		t.Logf("Fetching remote twtxt file ...\n")
		mainregistry, _, err := GetTwtxt("https://gbmor.dev/twtxt.txt", nil)
		if err != nil {
			t.Errorf("%v\n", err)
		}

		t.Logf("Parsing remote twtxt file ...\n")
		parsed, errz := ParseUserTwtxt(mainregistry, "gbmor", "https://gbmor.dev/twtxt.txt")
		if errz != nil {
			t.Errorf("%v\n", errz)
		}

		t.Logf("Adding new user to registry ...\n")
		err = registry.AddUser("TestRegistry", "https://gbmor.dev/twtxt.txt", nil, parsed)
		if err != nil {
			t.Errorf("%v\n", err)
		}

		t.Logf("Querying user statuses ...\n")
		queryuser, err := registry.QueryUser("TestRegistry")
		if err != nil {
			t.Errorf("%v\n", err)
		}
		for _, e := range queryuser {
			if !strings.Contains(e, "TestRegistry") {
				t.Errorf("QueryUser() returned incorrect data\n")
			}
		}

		t.Logf("Querying for keyword in statuses ...\n")
		querystatus, err := registry.QueryInStatus("morning")
		if err != nil {
			t.Errorf("%v\n", err)
		}
		for _, e := range querystatus {
			if !strings.Contains(e, "morning") {
				t.Errorf("QueryInStatus() returned incorrect data\n")
			}
		}

		t.Logf("Querying for all statuses ...\n")
		allstatus, err := registry.QueryAllStatuses()
		if err != nil {
			t.Errorf("%v\n", err)
		}
		if len(allstatus) == 0 || allstatus == nil {
			t.Errorf("Got nil/zero from QueryAllStatuses")
		}

		t.Logf("Querying for all users ...\n")
		allusers, err := registry.QueryUser("")
		if err != nil {
			t.Errorf("%v\n", err)
		}
		if len(allusers) == 0 || allusers == nil {
			t.Errorf("Got nil/zero users on empty QueryUser() query")
		}

		t.Logf("Deleting user ...\n")
		err = registry.DelUser("https://gbmor.dev/twtxt.txt")
		if err != nil {
			t.Errorf("%v\n", err)
		}
	}
	t.Run("Integration Test", integration)
}
