package main

import (
	"testing"
)

func Test_refreshCache(t *testing.T) {
	initTestConf()
	confObj.Mu.RLock()
	prevtime := confObj.LastCache
	confObj.Mu.RUnlock()

	t.Run("Cache Time Check", func(t *testing.T) {
		refreshCache()
		confObj.Mu.RLock()
		newtime := confObj.LastCache
		confObj.Mu.RUnlock()

		if !newtime.After(prevtime) || newtime == prevtime {
			t.Errorf("Cache time did not update, check refreshCache() logic\n")
		}
	})
}

func Benchmark_refreshCache(b *testing.B) {
	initTestConf()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		refreshCache()
	}
}
