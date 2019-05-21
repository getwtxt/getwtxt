package main

import "sync"

// content-type consts
const txtutf8 = "text/plain; charset=utf-8"
const htmlutf8 = "text/html; charset=utf-8"
const cssutf8 = "text/css; charset=utf-8"

// config object definition
type configuration struct {
	port          int
	logfile       string
	stdoutLogging bool
	version       string
	Instance
}

// Instance refers to this specific instance of getwtxt
type Instance struct {
	Name  string
	URL   string
	Owner string
	Mail  string
	Desc  string
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
