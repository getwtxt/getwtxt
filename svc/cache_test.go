package svc // import "github.com/getwtxt/getwtxt/svc"

import (
	"testing"
)

func Test_cacheUpdate(t *testing.T) {
	initTestConf()
	confObj.Mu.RLock()
	prevtime := confObj.LastCache
	confObj.Mu.RUnlock()

	t.Run("Cache Time Check", func(t *testing.T) {
		cacheUpdate()
		confObj.Mu.RLock()
		newtime := confObj.LastCache
		confObj.Mu.RUnlock()

		if !newtime.After(prevtime) || newtime == prevtime {
			t.Errorf("Cache time did not update, check cacheUpdate() logic\n")
		}
	})
}

func Benchmark_cacheUpdate(b *testing.B) {
	initTestConf()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cacheUpdate()
	}
}

func Benchmark_pingAssets(b *testing.B) {
	initTestConf()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		pingAssets()
	}
}
