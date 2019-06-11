package svc // import "github.com/getwtxt/getwtxt/svc"

import (
	"testing"

	"github.com/getwtxt/registry"
)

func Benchmark_cacheUpdate(b *testing.B) {
	initTestConf()
	mockRegistry()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cacheUpdate()

		// make sure it's pulling new statuses
		// half the time so we get a good idea
		// of its performance in both cases.
		if i > 2 && i%2 == 0 {
			b.StopTimer()
			twtxtCache.Mu.Lock()
			user := twtxtCache.Users["https://gbmor.dev/twtxt.txt"]
			user.Mu.Lock()
			user.Status = registry.NewTimeMap()
			user.RLen = "0"
			twtxtCache.Users["https://gbmor.dev/twtxt.txt"] = user
			user.Mu.Unlock()
			twtxtCache.Mu.Unlock()
			b.StartTimer()
		}
	}
}

func Benchmark_pingAssets(b *testing.B) {
	initTestConf()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		pingAssets()

		// We'll only have to reload the cache occasionally,
		// so only start with an empty staticCache 25% of
		// the time.
		if float64(i) > (float64(b.N) * .75) {
			b.StopTimer()
			staticCache = &staticAssets{}
			b.StartTimer()
		}
	}
}
