package svc // import "github.com/getwtxt/getwtxt/svc"

import (
	"net"
	"strings"
	"time"

	"github.com/getwtxt/registry"
	"github.com/syndtr/goleveldb/leveldb"
)

type dbLevel struct {
	db *leveldb.DB
}

func (lvl *dbLevel) push() error {
	twtxtCache.Mu.RLock()
	var dbBasket = &leveldb.Batch{}
	for k, v := range twtxtCache.Users {

		dbBasket.Put([]byte(k+"*Nick"), []byte(v.Nick))
		dbBasket.Put([]byte(k+"*URL"), []byte(v.URL))
		dbBasket.Put([]byte(k+"*IP"), []byte(v.IP.String()))
		dbBasket.Put([]byte(k+"*Date"), []byte(v.Date))
		dbBasket.Put([]byte(k+"*RLen"), []byte(v.RLen))

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

	return lvl.db.Write(dbBasket, nil)
}

func (lvl *dbLevel) pull() {
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
			twtxtCache.Users[urls].Mu.RLock()
			data = twtxtCache.Users[urls]
			twtxtCache.Users[urls].Mu.RUnlock()
		}
		twtxtCache.Mu.RUnlock()

		data.Mu.Lock()
		switch field {
		case "IP":
			data.IP = net.ParseIP(val)
		case "Nick":
			data.Nick = val
		case "URL":
			data.URL = val
		case "RLen":
			data.RLen = val
		case "Date":
			data.Date = val
		case "Status":
			thetime, err := time.Parse(time.RFC3339, split[2])
			errLog("", err)
			data.Status[thetime] = val
		}
		data.Mu.Unlock()

		twtxtCache.Mu.Lock()
		twtxtCache.Users[urls] = data
		twtxtCache.Mu.Unlock()
	}

	remoteRegistries.Mu.Lock()
	remoteRegistries.List = dedupe(remoteRegistries.List)
	remoteRegistries.Mu.Unlock()

	iter.Release()
	errLog("Error while pulling DB into registry cache: ", iter.Error())
}
