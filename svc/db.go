package svc // import "github.com/getwtxt/getwtxt/svc"

import (
	"time"

	"github.com/syndtr/goleveldb/leveldb"
	"golang.org/x/sys/unix"
)

type dbase interface {
	push() error
	pull()
}

// Pull DB data into cache, if available.
func initDatabase() {
	var db dbase
	confObj.Mu.RLock()
	dbpath := confObj.DBPath
	confObj.Mu.RUnlock()

	switch confObj.DBType {

	case "leveldb":
		lvl, err := leveldb.OpenFile(dbpath, nil)
		errFatal("", err)
		db = &dbLevel{db: lvl}

	case "sqlite":
		db = initSqlite()

	}

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

	unix.Sync()

	return err
}

func pullDB() {
	db := <-dbChan
	db.pull()
	dbChan <- db
}
