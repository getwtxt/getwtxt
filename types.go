package main

import "regexp"

const txtutf8 = "text/plain; charset=utf8"
const htmlutf8 = "text/html; charset=utf8"

type configuration struct {
	port         string
	validPath    *regexp.Regexp
	quietLogging bool
	fileLogging  bool
	logFile      string
}

var confObj = &configuration{}

func initConfig() {
	confObj.validPath = regexp.MustCompile("^/(api)/(plain)/(tweets|users|tags|mentions)")
}
