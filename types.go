package main

const txtutf8 = "text/plain; charset=utf8"
const htmlutf8 = "text/html; charset=utf8"

type configuration struct {
	port         string
	quietLogging bool
	fileLogging  bool
	logFile      string
}

var confObj = &configuration{}
