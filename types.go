package main

// content-type consts
const txtutf8 = "text/plain; charset=utf8"
const htmlutf8 = "text/html; charset=utf8"

// config object definition
type configuration struct {
	port         string
	quietLogging bool
	fileLogging  bool
	logFile      string
}

// config object
var confObj = &configuration{}
