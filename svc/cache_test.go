package svc // import "github.com/getwtxt/getwtxt/svc"

import (
	"testing"
	"time"
)

func Test_cacheTimer(t *testing.T) {
	initTestConf()
	dur, _ := time.ParseDuration("5m")
	back30, _ := time.ParseDuration("-30m")

	cases := []struct {
		name      string
		lastCache time.Time
		interval  time.Duration
		expect    bool
	}{
		{
			name:      "Past Interval",
			lastCache: time.Now().Add(back30),
			interval:  dur,
			expect:    true,
		},
		{
			name:      "Before Interval",
			lastCache: time.Now(),
			interval:  dur,
			expect:    false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {

			confObj.Mu.Lock()
			confObj.LastCache = tt.lastCache
			confObj.CacheInterval = tt.interval
			confObj.Mu.Unlock()

			res := cacheTimer()

			if res != tt.expect {
				t.Errorf("Got %v, expected %v\n", res, tt.expect)
			}
		})
	}

}

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

func Benchmark_pingAssets(b *testing.B) {
	initTestConf()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		pingAssets()
	}
}
