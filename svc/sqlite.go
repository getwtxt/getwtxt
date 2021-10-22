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
	"database/sql"
	"net"
	"time"

	"git.sr.ht/~gbmor/getwtxt/registry"
	_ "github.com/mattn/go-sqlite3" // for the sqlite3 driver
)

// Wrapper containing a SQLite database connection,
// along with two prepared statements for pushing
// and pulling via said connection.
type dbSqlite struct {
	db       *sql.DB
	pullStmt *sql.Stmt
	pushStmt *sql.Stmt
}

// Initializes a SQLite database.
func initSqlite() *dbSqlite {
	confObj.Mu.RLock()
	dbpath := confObj.DBPath
	confObj.Mu.RUnlock()

	lite, err := sql.Open("sqlite3", dbpath)
	errFatal("Error opening sqlite3 DB: ", err)

	errFatal("", lite.Ping())

	_, err = lite.Exec("CREATE TABLE IF NOT EXISTS getwtxt (id INTEGER PRIMARY KEY, urlKey TEXT, isUser BOOL, dataKey TEXT, data BLOB)")
	errFatal("Error preparing sqlite3 DB: ", err)

	push, err := lite.Prepare("INSERT OR REPLACE INTO getwtxt (urlKey, isUser, dataKey, data) VALUES(?, ?, ?, ?)")
	errFatal("", err)

	pull, err := lite.Prepare("SELECT * FROM getwtxt")
	errFatal("", err)

	return &dbSqlite{
		db:       lite,
		pushStmt: push,
		pullStmt: pull,
	}
}

func (lite *dbSqlite) delUser(userURL string) error {
	return nil
}

// Commits data from memory to a SQLite database intermittently.
func (lite *dbSqlite) push() error {
	if err := lite.db.Ping(); err != nil {
		lite = initSqlite()
	}

	tx, err := lite.db.Begin()
	errLog("", err)
	txst := tx.Stmt(lite.pushStmt)

	twtxtCache.Mu.RLock()
	defer twtxtCache.Mu.RUnlock()

	for i, e := range twtxtCache.Users {
		e.Mu.RLock()

		_, err = txst.Exec(i, true, "nickname", e.Nick)
		errLog("", err)
		_, err = txst.Exec(i, true, "lastmodified", e.LastModified)
		errLog("", err)
		_, err = txst.Exec(i, true, "uip", e.IP)
		errLog("", err)
		_, err = txst.Exec(i, true, "date", e.Date)
		errLog("", err)

		for k, v := range e.Status {
			_, err = txst.Exec(i, true, k.Format(time.RFC3339), v)
			errLog("", err)
		}
		e.Mu.RUnlock()
	}

	for _, e := range remoteRegistries.List {
		_, err = txst.Exec(e, false, "REMOTE REGISTRY", "NULL")
		errLog("", err)
	}

	err = tx.Commit()
	if err != nil {
		errLog("", tx.Rollback())
		return err
	}
	return nil
}

// Retrieves stored data from a SQLite database on startup.
func (lite *dbSqlite) pull() {
	errLog("Error pinging sqlite DB: ", lite.db.Ping())
	rows, err := lite.pullStmt.Query()
	errLog("", err)

	defer func(rows *sql.Rows) {
		errLog("Error while finalizing DB Pull: ", rows.Close())
	}(rows)

	twtxtCache.Mu.Lock()
	defer twtxtCache.Mu.Unlock()

	for rows.Next() {
		var uid int
		var urls string
		var isUser bool
		var dataKey string
		var dBlob []byte

		errLog("", rows.Scan(&uid, &urls, &isUser, &dataKey, &dBlob))
		if !isUser {
			remoteRegistries.List = append(remoteRegistries.List, urls)
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
		case "lastmodified":
			user.LastModified = string(dBlob)
		default:
			thetime, err := time.Parse(time.RFC3339, dataKey)
			errLog("While pulling statuses from SQLite: ", err)
			user.Status[thetime] = string(dBlob)
		}
		twtxtCache.Users[urls] = user
		user.Mu.Unlock()
	}
	remoteRegistries.List = dedupe(remoteRegistries.List)
}
