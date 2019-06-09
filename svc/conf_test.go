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
}
