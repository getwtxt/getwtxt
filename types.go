package main

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

// ipCtxKey is the Context value key for user IP addresses
type ipCtxKey int

const ctxKey ipCtxKey = iota
