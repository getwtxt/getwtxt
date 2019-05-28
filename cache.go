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

func checkCacheTime() bool {
	confObj.Mu.RLock()
	answer := time.Since(confObj.LastCache) > confObj.CacheInterval
	confObj.Mu.RUnlock()

	return answer
}

func checkDBtime() bool {
	confObj.Mu.RLock()
	answer := time.Since(confObj.LastPush) > confObj.DBInterval
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
				log.Printf("Error pushing cache to database: %v\n", err)
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
			log.Printf("%v\n", err)
		}
		twtxtCache.Mu.RLock()
	}
	twtxtCache.Mu.RUnlock()

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
	db := <-dbChan
	dbChan <- db

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

	remoteRegistries.Mu.RLock()
	for k, v := range remoteRegistries.List {
		dbBasket.Put([]byte("remote*"+string(k)), []byte(v))
	}
	remoteRegistries.Mu.RUnlock()

	if err := db.Write(dbBasket, nil); err != nil {
		return err
	}

	confObj.Mu.Lock()
	confObj.LastPush = time.Now()
	confObj.Mu.Unlock()

	return nil
}

func pullDatabase() {
	db := <-dbChan
	dbChan <- db

	iter := db.NewIterator(nil, nil)

	for iter.Next() {
		key := string(iter.Key())
		val := string(iter.Value())

		split := strings.Split(key, "*")
		urls := split[0]
		field := split[1]

		if urls == "remote" {
			remoteRegistries.Mu.Lock()
			remoteRegistries.List = append(remoteRegistries.List, val)
			remoteRegistries.Mu.Unlock()
			continue
		}

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
			thetime, err := time.Parse(time.RFC3339, split[2])
			if err != nil {
				log.Printf("%v\n", err)
			}
			data.Status[thetime] = val
		}

		twtxtCache.Mu.Lock()
		twtxtCache.Users[urls] = data
		twtxtCache.Mu.Unlock()

	}

	remoteRegistries.Mu.Lock()
	remoteRegistries.List = dedupe(remoteRegistries.List)
	remoteRegistries.Mu.Unlock()

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
