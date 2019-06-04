package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"time"
)

func checkCacheTime() bool {
	confObj.Mu.RLock()
	answer := time.Since(confObj.LastCache) > confObj.CacheInterval
	confObj.Mu.RUnlock()

	return answer
}

// Launched by init as a coroutine to watch
// for the update intervals to pass.
func cacheAndPush() {
	for {
		if checkCacheTime() {
			refreshCache()
		}
		if checkDBtime() {
			if err := pushDatabase(); err != nil {
				log.Printf("Error pushing cache to database: %v\n", err.Error())
			}
		}
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

	cssStat, err := os.Stat("assets/style.css")
	if err != nil {
		log.Printf("%v\n", err.Error())
	}

	indexStat, err := os.Stat("assets/tmpl/index.html")
	if err != nil {
		log.Printf("%v\n", err.Error())
	}

	indexMod := staticCache.indexMod
	cssMod := staticCache.cssMod

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

		staticCache.index = buf.Bytes()
		staticCache.indexMod = indexStat.ModTime()
	}

	if !cssMod.Equal(cssStat.ModTime()) {

		css, err := ioutil.ReadFile("assets/style.css")
		if err != nil {
			log.Printf("%v\n", err.Error())
		}

		staticCache.css = css
		staticCache.cssMod = cssStat.ModTime()
	}
}

// Simple function to deduplicate entries in a []string
func dedupe(list []string) []string {
	var out []string
	var seen = map[string]bool{}

	for _, e := range list {
		if !seen[e] {
			out = append(out, e)
			seen[e] = true
		}
	}

	return out
}
