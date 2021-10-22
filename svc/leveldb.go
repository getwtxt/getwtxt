/*
Copyright (c) 2019 Ben Morrison (gbmor)

This file is part of Getwtxt.

Getwtxt is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

Getwtxt is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with Getwtxt.  If not, see <https://www.gnu.org/licenses/>.
*/

package svc // import "git.sr.ht/~gbmor/getwtxt/svc"

import (
	"net"
	"strings"
	"time"

	"git.sr.ht/~gbmor/getwtxt/registry"
	"github.com/syndtr/goleveldb/leveldb"
)

// Wrapper type for the LevelDB connection
type dbLevel struct {
	db *leveldb.DB
}

func (lvl *dbLevel) delUser(userURL string) error {
	twtxtCache.Mu.RLock()
	defer twtxtCache.Mu.RUnlock()

	userStatuses := twtxtCache.Users[userURL].Status
	var dbBasket = &leveldb.Batch{}

	dbBasket.Delete([]byte(userURL + "*Nick"))
	dbBasket.Delete([]byte(userURL + "*URL"))
	dbBasket.Delete([]byte(userURL + "*IP"))
	dbBasket.Delete([]byte(userURL + "*Date"))
	dbBasket.Delete([]byte(userURL + "*LastModified"))

	for i := range userStatuses {
		rfc := i.Format(time.RFC3339)
		dbBasket.Delete([]byte(userURL + "*Status*" + rfc))
	}

	return lvl.db.Write(dbBasket, nil)
}

// Called intermittently to commit registry data to
// a LevelDB database.
func (lvl *dbLevel) push() error {
	twtxtCache.Mu.RLock()
	defer twtxtCache.Mu.RUnlock()

	var dbBasket = &leveldb.Batch{}
	for k, v := range twtxtCache.Users {
		dbBasket.Put([]byte(k+"*Nick"), []byte(v.Nick))
		dbBasket.Put([]byte(k+"*URL"), []byte(v.URL))
		dbBasket.Put([]byte(k+"*IP"), []byte(v.IP.String()))
		dbBasket.Put([]byte(k+"*Date"), []byte(v.Date))
		dbBasket.Put([]byte(k+"*LastModified"), []byte(v.LastModified))

		for i, e := range v.Status {
			rfc := i.Format(time.RFC3339)
			dbBasket.Put([]byte(k+"*Status*"+rfc), []byte(e))
		}
	}

	for k, v := range remoteRegistries.List {
		dbBasket.Put([]byte("remote*"+string(k)), []byte(v))
	}

	return lvl.db.Write(dbBasket, nil)
}

// Called on startup to retrieve previously-committed data
// from a LevelDB database. Stores the retrieved data in
// memory.
func (lvl *dbLevel) pull() {
	iter := lvl.db.NewIterator(nil, nil)
	twtxtCache.Mu.Lock()
	defer twtxtCache.Mu.Unlock()

	for iter.Next() {
		key := string(iter.Key())
		val := string(iter.Value())
		split := strings.Split(key, "*")
		urls := split[0]
		field := split[1]

		if urls == "remote" {
			remoteRegistries.List = append(remoteRegistries.List, val)
			continue
		}

		data := registry.NewUser()
		if _, ok := twtxtCache.Users[urls]; ok {
			data = twtxtCache.Users[urls]
		}

		data.Mu.Lock()
		switch field {
		case "IP":
			data.IP = net.ParseIP(val)
		case "Nick":
			data.Nick = val
		case "URL":
			data.URL = val
		case "LastModified":
			data.LastModified = val
		case "Date":
			data.Date = val
		case "Status":
			thetime, err := time.Parse(time.RFC3339, split[2])
			errLog("", err)
			data.Status[thetime] = val
		}
		twtxtCache.Users[urls] = data
		data.Mu.Unlock()
	}

	remoteRegistries.List = dedupe(remoteRegistries.List)
	iter.Release()
	errLog("Error while pulling DB into registry cache: ", iter.Error())
}
