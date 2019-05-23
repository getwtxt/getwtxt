package main

import (
	"sync"
	"time"
)

// content-type consts
const txtutf8 = "text/plain; charset=utf-8"
const htmlutf8 = "text/html; charset=utf-8"
const cssutf8 = "text/css; charset=utf-8"

// Configuration object definition
type Configuration struct {
	Mu            sync.RWMutex
	Port          int           `json:"ListenPort"`
	LogFile       string        `json:"LogFile"`
	DBPath        string        `json:"DatabasePath"`
	StdoutLogging bool          `json:"StdoutLogging"`
	Version       string        `json:"-"`
	CacheInterval time.Duration `json:"StatusFetchInterval"`
	DBInterval    time.Duration `json:"DatabasePushInterval"`
	LastCache     time.Time     `json:"-"`
	LastPush      time.Time     `json:"-"`
	Instance      `json:"Instance"`
}

// Instance refers to this specific instance of getwtxt
type Instance struct {
	Name  string `json:"Instance.SiteName"`
	URL   string `json:"Instance.URL"`
	Owner string `json:"Instance.OwnerName"`
	Mail  string `json:"Instance.Email"`
	Desc  string `json:"Instance.Description"`
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
