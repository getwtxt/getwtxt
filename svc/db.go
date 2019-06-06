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
	db       *sql.DB
	pullStmt *sql.Stmt
	pushStmt *sql.Stmt
}

// Pull DB data into cache, if available.
func initDatabase() {
	var db dbase

	confObj.Mu.RLock()
	switch confObj.DBType {

	case "leveldb":
		lvl, err := leveldb.OpenFile(confObj.DBPath, nil)
		errFatal("", err)
		db = &dbLevel{db: lvl}

	case "sqlite":
		db = initSqlite()

	}
	confObj.Mu.RUnlock()

	dbChan <- db
	pullDB()
}

func dbTimer() bool {
	confObj.Mu.RLock()
	answer := time.Since(confObj.LastPush) > confObj.DBInterval
	confObj.Mu.RUnlock()

	return answer
}

func initSqlite() *dbSqlite {

	lite, err := sql.Open("sqlite3", confObj.DBPath)
	errFatal("Error opening sqlite3 DB: ", err)

	litePrep, err := lite.Prepare("CREATE TABLE IF NOT EXISTS getwtxt (urlKey TEXT, isUser BOOL, dataKey TEXT, data BLOB)")
	errFatal("Error preparing sqlite3 DB: ", err)

	_, err = litePrep.Exec()
	errFatal("Error creating sqlite3 DB: ", err)

	push, err := lite.Prepare("INSERT OR REPLACE INTO getwtxt(urlKey, isUser, dataKey, data) VALUES(?, ?, ?, ?)")
	errFatal("", err)

	pull, err := lite.Prepare("SELECT * FROM getwtxt")
	errFatal("", err)

	return &dbSqlite{
		db:       lite,
		pushStmt: push,
		pullStmt: pull,
	}
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

	return lvl.db.Write(dbBasket, nil)
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

func (lite dbSqlite) push() error {
	err := lite.db.Ping()
	if err != nil {
		return err
	}

	twtxtCache.Mu.RLock()
	for i, e := range twtxtCache.Users {
		e.Mu.RLock()

		_, err = lite.pushStmt.Exec(i, true, "nickname", e.Nick)
		errLog("", err)
		_, err = lite.pushStmt.Exec(i, true, "rlen", e.RLen)
		errLog("", err)
		_, err = lite.pushStmt.Exec(i, true, "uip", e.IP)
		errLog("", err)
		_, err = lite.pushStmt.Exec(i, true, "date", e.Date)
		errLog("", err)

		for k, v := range e.Status {
			_, err = lite.pushStmt.Exec(i, true, k.Format(time.RFC3339), v)
			errLog("", err)
		}

		e.Mu.RUnlock()
	}
	twtxtCache.Mu.RUnlock()

	remoteRegistries.Mu.RLock()
	for _, e := range remoteRegistries.List {
		_, err = lite.pushStmt.Exec(e, false, "REMOTE REGISTRY", "NULL")
		errLog("", err)
	}
	remoteRegistries.Mu.RUnlock()

	return nil
}

func (lite dbSqlite) pull() {
	errLog("Error pinging sqlite DB: ", lite.db.Ping())

	rows, err := lite.pullStmt.Query()
	errLog("", err)

	twtxtCache.Mu.Lock()
	for rows.Next() {
		var urls string
		var isUser bool
		var dataKey string
		var dBlob []byte

		errLog("", rows.Scan(&urls, &isUser, &dataKey, &dBlob))

		if !isUser {
			remoteRegistries.Mu.Lock()
			remoteRegistries.List = append(remoteRegistries.List, urls)
			remoteRegistries.Mu.Unlock()
			continue
		}

		user := registry.NewUser()
		if _, ok := twtxtCache.Users[urls]; ok {
			user = twtxtCache.Users[urls]
		}
		user.Mu.Lock()

		switch dataKey {
		case "nickname":
			user.Nick = string(dBlob)
		case "uip":
			user.IP = net.ParseIP(string(dBlob))
		case "date":
			user.Date = string(dBlob)
		case "rlen":
			user.RLen = string(dBlob)
		default:
			thetime, err := time.Parse(time.RFC3339, dataKey)
			errLog("While pulling statuses from SQLite: ", err)
			user.Status[thetime] = string(dBlob)
		}

		twtxtCache.Users[urls] = user
		user.Mu.Unlock()
	}
	twtxtCache.Mu.Unlock()

	remoteRegistries.Mu.Lock()
	remoteRegistries.List = dedupe(remoteRegistries.List)
	remoteRegistries.Mu.Unlock()
}
