package main

import (
	"fmt"
	"log"
	"os"
)

var testport = fmt.Sprintf(":%v", confObj.Port)

func initTestConf() {
	initConfig()
	tmpls = initTemplates()
	logToNull()
}

func logToNull() {
	hush, err := os.Open("/dev/null")
	if err != nil {
		log.Printf("%v\n", err)
	}
	log.SetOutput(hush)
}
