package main

import (
	"database/sql"
	"sync"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
)

// content-type consts
const txtutf8 = "text/plain; charset=utf-8"
const htmlutf8 = "text/html; charset=utf-8"
const cssutf8 = "text/css; charset=utf-8"

// Configuration object definition
type Configuration struct {
	Mu            sync.RWMutex
	Port          int           `yaml:"ListenPort"`
	LogFile       string        `yaml:"LogFile"`
	DBType        string        `yaml:"DatabaseType"`
	DBPath        string        `yaml:"DatabasePath"`
	AssetsDir     string        `yaml:"-"`
	StdoutLogging bool          `yaml:"StdoutLogging"`
	Version       string        `yaml:"-"`
	CacheInterval time.Duration `yaml:"StatusFetchInterval"`
	DBInterval    time.Duration `yaml:"DatabasePushInterval"`
	LastCache     time.Time     `yaml:"-"`
	LastPush      time.Time     `yaml:"-"`
	Instance      `yaml:"Instance"`
}

// Instance refers to this specific instance of getwtxt
type Instance struct {
	Vers  string `yaml:"-"`
	Name  string `yaml:"Instance.SiteName"`
	URL   string `yaml:"Instance.URL"`
	Owner string `yaml:"Instance.OwnerName"`
	Mail  string `yaml:"Instance.Email"`
	Desc  string `yaml:"Instance.Description"`
}

type dbLevel struct {
	db *leveldb.DB
}

type dbSqlite struct {
	db *sql.DB
}

type dbase interface {
	push() error
	pull()
}

// RemoteRegistries holds a list of remote registries to
// periodically scrape for new users. The remote registries
// must have been added via POST like a user.
type RemoteRegistries struct {
	Mu   sync.RWMutex
	List []string
}

// ipCtxKey is the Context value key for user IP addresses
type ipCtxKey int

const ctxKey ipCtxKey = iota
