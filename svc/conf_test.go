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

package svc

import (
	"reflect"
	"testing"
)

func Test_initConfig(t *testing.T) {
	t.Run("Testing Configuration Initialization", func(t *testing.T) {

		initConfig()
		confObj.Mu.RLock()
		cnf := reflect.Indirect(reflect.ValueOf(confObj))
		confObj.Mu.RUnlock()

		for i := 0; i < cnf.NumField(); i++ {
			if !cnf.Field(i).IsValid() {
				t.Errorf("Uninitialized value: %v\n", cnf.Field(i).Type())
			}
		}
	})
	testConfig()
}
