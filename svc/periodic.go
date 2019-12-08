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

package svc // import "github.com/getwtxt/getwtxt/svc"

import (
	"log"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Functions and types in this file pertain
// to periodic, regular actions.

// This is a wrapper for a *time.Ticker
// that adds another channel. It's used
// to signal to the ticker goroutines
// that they should stop the tickers
// and exit.
type tick struct {
	isDB bool
	t    *time.Ticker
	exit chan struct{}
}

// Creates a new instance of a tick
func initTicker(db bool, interval time.Duration) *tick {
	return &tick{
		isDB: db,
		t:    time.NewTicker(interval),
		exit: make(chan struct{}, 1),
	}
}

// Sends the signal to stop the tickers
// and for their respective goroutines
// to exit.
func killTickers() {
	ct := <-cTickC
	dt := <-dbTickC
	ct.exit <- struct{}{}
	dt.exit <- struct{}{}
}

// Waits for a signal from the database
// *tick. Either stops the ticker and
// kills the goroutine or it will
// update cache / push the DB to disk
func dataTimer(tkr *tick) {
	for {
		select {
		case signal := <-tkr.t.C:
			if tkr.isDB {
				errLog("", pushDB())
				log.Printf("Database push took: %v\n", time.Since(signal))
				continue
			}
			cacheUpdate()
			log.Printf("Cache update took: %v\n", time.Since(signal))
		case <-tkr.exit:
			tkr.t.Stop()
			return
		}
	}
}

// Called when a change is detected in the
// configuration file. Closes log file,
// closes database connection, stops all
// tickers, then binds new configuration
// values, opens new log file, connects to
// new database, and starts new cache and
// database tickers.
func reInit(e fsnotify.Event) {
	log.Printf("%v. Reloading...\n", e.String())

	if !confObj.StdoutLogging {
		closeLog <- struct{}{}
	}

	killTickers()
	killDB()

	bindConfig()

	initLogging()
	initDatabase()
	initPersistence()
}

// Starts the tickers that periodically:
//  - pull new user statuses into cache
//  - push cached data to disk
func initPersistence() {
	confObj.Mu.RLock()
	cacheTkr := initTicker(false, confObj.CacheInterval)
	dbTkr := initTicker(true, confObj.DBInterval)
	confObj.Mu.RUnlock()

	go dataTimer(cacheTkr)
	go dataTimer(dbTkr)

	dbTickC <- dbTkr
	cTickC <- cacheTkr
}
