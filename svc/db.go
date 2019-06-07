package svc // import "github.com/getwtxt/getwtxt/svc"

import (
	"database/sql"
	"net"
	"strings"
	"time"

	"github.com/getwtxt/registry"
	_ "github.com/mattn/go-sqlite3" // for the sqlite3 driver
	"github.com/syndtr/goleveldb/leveldb"
)

type dbase interface {
	push() error
	pull()
}

type dbLevel struct {
	db *leveldb.DB
}

type dbSqlite struct {
	db *sql.DB
}

type dbPostgres struct {
	db *sql.DB
}

// Pull DB data into cache, if available.
func initDatabase() {
	var db dbase
	var err error

	confObj.Mu.RLock()
	switch confObj.DBType {

	case "leveldb":
		var lvl *leveldb.DB
		lvl, err = leveldb.OpenFile(confObj.DBPath, nil)
		db = &dbLevel{db: lvl}

	case "sqlite":
		var lite *sql.DB
		lite, err := sql.Open("sqlite3", confObj.DBPath)
		errFatal("Error opening sqlite3 DB: ", err)
		litePrep, err := lite.Prepare("CREATE TABLE IF NOT EXISTS getwtxt (urlKey TEXT PRIMARY KEY, isUser BOOL, blobKey TEXT, data BLOB)")
		errFatal("Error preparing sqlite3 DB: ", err)
		_, err = litePrep.Exec()
		errFatal("Error creating sqlite3 DB: ", err)
		db = &dbSqlite{db: lite}

	case "postgres":
		var pg *sql.DB
		db = &dbPostgres{db: pg}

	}
	confObj.Mu.RUnlock()

	errFatal("", err)

	dbChan <- db

	pullDB()
}

func dbTimer() bool {
	confObj.Mu.RLock()
	answer := time.Since(confObj.LastPush) > confObj.DBInterval
	confObj.Mu.RUnlock()

	return answer
}

// Pushes the registry's cache data to a local
// database for safe keeping.
func pushDB() error {
	db := <-dbChan
	err := db.push()
	dbChan <- db

	return err
}

func pullDB() {
	db := <-dbChan
	db.pull()
	dbChan <- db
}

func (lvl dbLevel) push() error {
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

	confObj.Mu.Lock()
	confObj.LastPush = time.Now()
	confObj.Mu.Unlock()

	err := lvl.db.Write(dbBasket, nil)

	return err
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
		case "RLen":
			data.RLen = val
		case "Date":
			data.Date = val
		case "Status":
			thetime, err := time.Parse(time.RFC3339, split[2])
			errLog("", err)
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
	errLog("Error while pulling DB into registry cache: ", err)
}

func (lite dbSqlite) push() error {

	return nil
}

func (lite dbSqlite) pull() {

}

func (pg dbPostgres) push() error {

	return nil
}

func (pg dbPostgres) pull() {

}
