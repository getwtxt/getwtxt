package svc // import "github.com/getwtxt/getwtxt/svc"

import (
	"testing"
)

func Benchmark_cacheUpdate(b *testing.B) {
	initTestConf()
	mockRegistry()
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
