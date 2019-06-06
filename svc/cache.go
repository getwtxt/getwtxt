package svc // import "github.com/getwtxt/getwtxt/svc"

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"
)

// RemoteRegistries holds a list of remote registries to
// periodically scrape for new users. The remote registries
// must have been added via POST like a user.
type RemoteRegistries struct {
	Mu   sync.RWMutex
	List []string
}

type staticAssets struct {
	mu       sync.RWMutex
	index    []byte
	indexMod time.Time
	css      []byte
	cssMod   time.Time
}

func cacheTimer() bool {
	confObj.Mu.RLock()
	answer := time.Since(confObj.LastCache) > confObj.CacheInterval
	confObj.Mu.RUnlock()

	return answer
}

// Launched by init as a coroutine to watch
// for the update intervals to pass.
func cacheAndPush() {
	for {
		if cacheTimer() {
			refreshCache()
		}
		if dbTimer() {
			if err := pushDB(); err != nil {
				log.Printf("Error pushing cache to database: %v\n", err.Error())
			}
		}
		time.Sleep(1000 * time.Millisecond)
	}
}

func refreshCache() {

	// This clusterfuck of mutex read locks is
	// necessary to avoid deadlock. This mess
	// also avoids a panic that would occur
	// should twtxtCache be written to during
	// this loop.
	twtxtCache.Mu.RLock()
	for k := range twtxtCache.Users {
		twtxtCache.Mu.RUnlock()
		err := twtxtCache.UpdateUser(k)
		if err != nil {
			log.Printf("%v\n", err.Error())
		}
		twtxtCache.Mu.RLock()
	}
	twtxtCache.Mu.RUnlock()

	remoteRegistries.Mu.RLock()
	for _, v := range remoteRegistries.List {
		err := twtxtCache.CrawlRemoteRegistry(v)
		if err != nil {
			log.Printf("Error while refreshing local copy of remote registry user data: %v\n", err.Error())
		}
	}
	remoteRegistries.Mu.RUnlock()
	confObj.Mu.Lock()
	confObj.LastCache = time.Now()
	confObj.Mu.Unlock()
}

// pingAssets checks if the local static assets
// need to be re-cached. If they do, they are
// pulled back into memory from disk.
func pingAssets() {

	confObj.Mu.RLock()
	assetsDir := confObj.AssetsDir
	confObj.Mu.RUnlock()

	cssStat, err := os.Stat(assetsDir + "/style.css")
	if err != nil {
		log.Printf("%v\n", err.Error())
	}

	indexStat, err := os.Stat(assetsDir + "/tmpl/index.html")
	if err != nil {
		log.Printf("%v\n", err.Error())
	}

	staticCache.mu.RLock()
	indexMod := staticCache.indexMod
	cssMod := staticCache.cssMod
	staticCache.mu.RUnlock()

	if !indexMod.Equal(indexStat.ModTime()) {
		tmpls = initTemplates()

		var b []byte
		buf := bytes.NewBuffer(b)

		confObj.Mu.RLock()
		err = tmpls.ExecuteTemplate(buf, "index.html", confObj.Instance)
		confObj.Mu.RUnlock()
		if err != nil {
			log.Printf("%v\n", err.Error())
		}

		staticCache.mu.Lock()
		staticCache.index = buf.Bytes()
		staticCache.indexMod = indexStat.ModTime()
		staticCache.mu.Unlock()
	}

	if !cssMod.Equal(cssStat.ModTime()) {

		css, err := ioutil.ReadFile(assetsDir + "/style.css")
		if err != nil {
			log.Printf("%v\n", err.Error())
		}

		staticCache.mu.Lock()
		staticCache.css = css
		staticCache.cssMod = cssStat.ModTime()
		staticCache.mu.Unlock()
	}
}
