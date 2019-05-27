package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/getwtxt/registry"
	"github.com/syndtr/goleveldb/leveldb"
)

// Checks whether it's time to refresh
// the cache.
func checkCacheTime() bool {
	return time.Since(confObj.LastCache) > confObj.CacheInterval
}

// Checks whether it's time to push
// the cache to the database
func checkDBtime() bool {
	return time.Since(confObj.LastPush) > confObj.DBInterval
}

// Launched by init as a goroutine to constantly watch
// for the update interval to pass.
func cacheAndPush() {
	for {
		if checkCacheTime() {
			refreshCache()
		}
		if checkDBtime() {
			if err := pushDatabase(); err != nil {
				log.Printf("Error pushing cache to database: %v\n", err)
			}
		}
	}
}

// Refreshes the cache.
func refreshCache() {

	// Iterate over the registry and
	// update each individual user.
	for k := range twtxtCache.Users {
		err := twtxtCache.UpdateUser(k)
		if err != nil {
			log.Printf("%v\n", err)
			continue
		}
	}

	// Re-scrape all the remote registries
	// to see if they have any new users
	// to add locally.
	remoteRegistries.Mu.RLock()
	for _, v := range remoteRegistries.List {
		err := twtxtCache.CrawlRemoteRegistry(v)
		if err != nil {
			log.Printf("Error while refreshing local copy of remote registry user data: %v\n", err)
		}
	}
	remoteRegistries.Mu.RUnlock()
	confObj.Mu.Lock()
	confObj.LastCache = time.Now()
	confObj.Mu.Unlock()
}

// Pushes the registry's cache data to a local
// database for safe keeping.
func pushDatabase() error {
	// Acquire the database from the aether.
	// goleveldb is concurrency-safe, so we
	// can immediately push it back into the
	// channel for other functions to use.
	db := <-dbChan
	dbChan <- db

	// Create a batch write job so it can
	// be done at one time rather than
	// per entry.
	twtxtCache.Mu.RLock()
	var dbBasket = &leveldb.Batch{}
	for k, v := range twtxtCache.Users {
		dbBasket.Put([]byte(k+"*Nick"), []byte(v.Nick))
		dbBasket.Put([]byte(k+"*URL"), []byte(v.URL))
		dbBasket.Put([]byte(k+"*IP"), []byte(v.IP.String()))
		dbBasket.Put([]byte(k+"*Date"), []byte(v.Date))
		for i, e := range v.Status {
			rfc := i.Format(time.RFC3339)
			dbBasket.Put([]byte(k+"*Status*"+rfc), []byte(e))
		}
	}
	twtxtCache.Mu.RUnlock()

	// Save our list of remote registries to scrape.
	remoteRegistries.Mu.RLock()
	for k, v := range remoteRegistries.List {
		dbBasket.Put([]byte("remote*"+string(k)), []byte(v))
	}
	remoteRegistries.Mu.RUnlock()

	// Execute the batch job.
	if err := db.Write(dbBasket, nil); err != nil {
		return err
	}

	// Update the last push time for
	// our timer/watch function to
	// reference.
	confObj.Mu.Lock()
	confObj.LastPush = time.Now()
	confObj.Mu.Unlock()

	return nil
}

// Pulls registry data from the DB on startup.
// Iterates over the database one entry at a time.
func pullDatabase() {
	// Acquire the database from the aether.
	// goleveldb is concurrency-safe, so we
	// can immediately push it back into the
	// channel for other functions to use.
	db := <-dbChan
	dbChan <- db

	iter := db.NewIterator(nil, nil)

	// Read the database entry-by-entry
	for iter.Next() {
		key := string(iter.Key())
		val := string(iter.Value())

		split := strings.Split(key, "*")
		urls := split[0]
		field := split[1]

		if urls != "remote" {
			// Start with an empty Data struct. If
			// there's already one in the cache, pull
			// it and use it instead.
			data := registry.NewUser()
			twtxtCache.Mu.RLock()
			if _, ok := twtxtCache.Users[urls]; ok {
				data = twtxtCache.Users[urls]
			}
			twtxtCache.Mu.RUnlock()

			switch field {
			case "IP":
				data.IP = net.ParseIP(val)
			case "Nick":
				data.Nick = val
			case "URL":
				data.URL = val
			case "Date":
				data.Date = val
			case "Status":
				// If we're looking at a Status entry in the DB,
				// parse the time then add it to the TimeMap under
				// data.Status
				thetime, err := time.Parse(time.RFC3339, split[2])
				if err != nil {
					log.Printf("%v\n", err)
				}
				data.Status[thetime] = val
			}

			// Push the data struct (back) into
			// the cache.
			twtxtCache.Mu.Lock()
			twtxtCache.Users[urls] = data
			twtxtCache.Mu.Unlock()

		} else {
			// If we've come across an entry for
			// a remote twtxt registry to scrape,
			// add it to our list.
			remoteRegistries.Mu.Lock()
			remoteRegistries.List = append(remoteRegistries.List, val)
			remoteRegistries.Mu.Unlock()
		}
	}

	iter.Release()
	err := iter.Error()
	if err != nil {
		log.Printf("Error while pulling DB into registry cache: %v\n", err)
	}
}

// pingAssets checks if the local static assets
// need to be re-cached. If they do, they are
// pulled back into memory from disk.
func pingAssets() {

	cssStat, err := os.Stat("assets/style.css")
	if err != nil {
		log.Printf("%v\n", err)
	}

	indexStat, err := os.Stat("assets/tmpl/index.html")
	if err != nil {
		log.Printf("%v\n", err)
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
			log.Printf("%v\n", err)
		}

		staticCache.index = buf.Bytes()
		staticCache.indexMod = indexStat.ModTime()
	}

	if !cssMod.Equal(cssStat.ModTime()) {

		css, err := ioutil.ReadFile("assets/style.css")
		if err != nil {
			log.Printf("%v\n", err)
		}

		staticCache.css = css
		staticCache.cssMod = cssStat.ModTime()
	}
}
