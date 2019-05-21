package main

import (
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/getwtxt/registry"
	"github.com/syndtr/goleveldb/leveldb"
)

// checks if it's time to refresh the cache or not
func checkCacheTime() bool {
	return time.Since(confObj.lastCache) > confObj.cacheInterval
}

// checks if it's time to push the cache to the database
func checkDBtime() bool {
	return time.Since(confObj.lastPush) > confObj.dbInterval
}

// launched by init as a goroutine to constantly watch
// for the update interval to pass
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

// refreshes the cache
func refreshCache() {

	for k := range twtxtCache.Reg {
		err := twtxtCache.UpdateUser(k)
		if err != nil {
			log.Printf("%v\n", err)
			continue
		}
	}

	for _, v := range remoteRegistries.List {
		err := twtxtCache.ScrapeRemoteRegistry(v)
		if err != nil {
			log.Printf("Error while refreshing local copy of remote registry user data: %v\n", err)
		}
	}
	confObj.mu.Lock()
	confObj.lastCache = time.Now()
	confObj.mu.Unlock()
}

// pushes the registry's cache data to a local
// database for safe keeping
func pushDatabase() error {
	db := <-dbChan
	twtxtCache.Mu.RLock()

	// create a batch write job so it can
	// be done at one time rather than
	// per value
	var dbBasket *leveldb.Batch
	for k, v := range twtxtCache.Reg {
		dbBasket.Put([]byte(k+".Nick"), []byte(v.Nick))
		dbBasket.Put([]byte(k+".URL"), []byte(v.URL))
		dbBasket.Put([]byte(k+".IP"), []byte(v.IP))
		dbBasket.Put([]byte(k+".Date"), []byte(v.Date))
		for i, e := range v.Status {
			dbBasket.Put([]byte(k+".Status."+i.String()), []byte(e))
		}
	}

	// save our list of remote registries to scrape
	for k, v := range remoteRegistries.List {
		dbBasket.Put([]byte("remote."+string(k)), []byte(v))
	}

	// execute the batch job
	if err := db.Write(dbBasket, nil); err != nil {
		return err
	}

	twtxtCache.Mu.RUnlock()
	dbChan <- db

	// update the last push time
	confObj.mu.Lock()
	confObj.lastPush = time.Now()
	confObj.mu.Unlock()

	return nil
}

// pulls registry data from the DB on startup
func pullDatabase() {
	db := <-dbChan

	iter := db.NewIterator(nil, nil)

	for iter.Next() {
		key := iter.Key()
		val := iter.Value()

		split := strings.Split(string(key), ".")
		urls := string(split[0])
		field := string(split[1])
		data := registry.NewUserData()

		twtxtCache.Mu.RLock()
		if _, ok := twtxtCache.Reg[urls]; ok {
			data = twtxtCache.Reg[urls]
		}
		twtxtCache.Mu.RUnlock()

		ref := reflect.ValueOf(data).Elem()

		if field != "Status" && urls != "remote" {
			for i := 0; i < ref.NumField(); i++ {

				f := ref.Field(i)
				if f.String() == field {
					f.Set(reflect.ValueOf(val))
					break
				}
			}
		} else if field == "Status" && urls != "remote" {

			thetime, err := time.Parse("RFC3339", split[2])
			if err != nil {
				log.Printf("%v\n", err)
			}
			data.Status[thetime] = string(val)

		} else {
			remoteRegistries.Mu.Lock()
			remoteRegistries.List = append(remoteRegistries.List, string(val))
			remoteRegistries.Mu.Unlock()
		}

		twtxtCache.Mu.Lock()
		twtxtCache.Reg[urls] = data
		twtxtCache.Mu.Unlock()
	}

	iter.Release()
	err := iter.Error()
	if err != nil {
		log.Printf("Error while pulling DB into registry cache: %v\n", err)
	}

	dbChan <- db
}
