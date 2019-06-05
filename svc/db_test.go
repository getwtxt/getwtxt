package svc // import "github.com/getwtxt/getwtxt/svc"

import (
	"net"
	"testing"

	"github.com/getwtxt/registry"
)

/*
func Test_pushpullDatabase(t *testing.T) {
	initTestConf()
	initDatabase()
	out, _, err := registry.GetTwtxt("https://gbmor.dev/twtxt.txt")
	if err != nil {
		t.Errorf("Couldn't set up test: %v\n", err)
	}
	statusmap, err := registry.ParseUserTwtxt(out, "gbmor", "https://gbmor.dev/twtxt.txt")
	if err != nil {
		t.Errorf("Couldn't set up test: %v\n", err)
	}
	twtxtCache.AddUser("gbmor", "https://gbmor.dev/twtxt.txt", "", net.ParseIP("127.0.0.1"), statusmap)
	remoteRegistries.Mu.Lock()
	remoteRegistries.List = append(remoteRegistries.List, "https://twtxt.tilde.institute/api/plain/users")
	remoteRegistries.Mu.Unlock()

	t.Run("Push to Database", func(t *testing.T) {
		err := pushDatabase()
		if err != nil {
			t.Errorf("%v\n", err)
		}
	})

	t.Run("Clearing Registry", func(t *testing.T) {
		err := twtxtCache.DelUser("https://gbmor.dev/twtxt.txt")
		if err != nil {
			t.Errorf("%v", err)
		}
	})

	t.Run("Pulling from Database", func(t *testing.T) {
		pullDatabase()
		twtxtCache.Mu.RLock()
		if _, ok := twtxtCache.Users["https://gbmor.dev/twtxt.txt"]; !ok {
			t.Errorf("Missing user previously pushed to database\n")
		}
		twtxtCache.Mu.RUnlock()

	})
}
*/
func Benchmark_pushDatabase(b *testing.B) {
	initTestConf()

	if len(dbChan) < 1 {
		initDatabase()
	}

	if _, ok := twtxtCache.Users["https://gbmor.dev/twtxt.txt"]; !ok {
		out, _, err := registry.GetTwtxt("https://gbmor.dev/twtxt.txt")
		if err != nil {
			b.Errorf("Couldn't set up benchmark: %v\n", err)
		}

		statusmap, err := registry.ParseUserTwtxt(out, "gbmor", "https://gbmor.dev/twtxt.txt")
		if err != nil {
			b.Errorf("Couldn't set up benchmark: %v\n", err)
		}

		twtxtCache.AddUser("gbmor", "https://gbmor.dev/twtxt.txt", "", net.ParseIP("127.0.0.1"), statusmap)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := pushDatabase()
		if err != nil {
			b.Errorf("%v\n", err)
		}
	}
}
func Benchmark_pullDatabase(b *testing.B) {
	initTestConf()

	if len(dbChan) < 1 {
		initDatabase()
	}

	for i := 0; i < b.N; i++ {
		pullDatabase()
	}
}
