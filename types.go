package main

import "regexp"

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
