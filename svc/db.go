package svc // import "github.com/getwtxt/getwtxt/svc"

import (
	"log"
	"net"
	"strings"
	"time"

	"github.com/getwtxt/registry"
	"github.com/syndtr/goleveldb/leveldb"
)

func checkDBtime() bool {
	confObj.Mu.RLock()
	answer := time.Since(confObj.LastPush) > confObj.DBInterval
	confObj.Mu.RUnlock()

	return answer
}

// Pushes the registry's cache data to a local
// database for safe keeping.
func pushDatabase() error {
	db := <-dbChan
	dbChan <- db

	return db.push()
}

func pullDatabase() {
	db := <-dbChan
	dbChan <- db
	db.pull()
}

func (lvl dbLevel) push() error {
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

	if err := lvl.db.Write(dbBasket, nil); err != nil {
		return err
	}

	confObj.Mu.Lock()
	confObj.LastPush = time.Now()
	confObj.Mu.Unlock()

	return nil
}

func (lvl dbLevel) pull() {

	iter := lvl.db.NewIterator(nil, nil)

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
				log.Printf("%v\n", err.Error())
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
		log.Printf("Error while pulling DB into registry cache: %v\n", err.Error())
	}
}

func (lite dbSqlite) push() error {

	return nil
}

func (lite dbSqlite) pull() {

}
