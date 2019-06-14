package svc // import "github.com/getwtxt/getwtxt/svc"

import (
	"log"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
	"golang.org/x/sys/unix"
)

// Everything in this file is database-agnostic.
// Functions and types related to specific kinds
// of databases will be in their own respective
// files, such as:
//  - leveldb.go
//  - sqlite.go

// Abstraction to allow several different
// databases to be used interchangeably.
type dbase interface {
	push() error
	pull()
}

// Opens a new connection to the specified
// database, then reads it into memory.
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

// Close the database connection.
func killDB() {
	db := <-dbChan
	switch dbType := db.(type) {
	case *dbLevel:
		errLog("", dbType.db.Close())
	case *dbSqlite:
		errLog("", dbType.db.Close())
	}
}

// Pushes the registry's cache data
// to a local database for safe keeping.
func pushDB() error {
	db := <-dbChan
	err := db.push()
	dbChan <- db

	unix.Sync()

	return err
}

// Reads the database from disk into memory.
func pullDB() {
	start := time.Now()
	db := <-dbChan
	db.pull()
	dbChan <- db
	log.Printf("Database pull took: %v\n", time.Since(start))
}
