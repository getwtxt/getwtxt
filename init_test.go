package main

import (
	"fmt"
	"log"
	"os"
)

var testport = fmt.Sprintf(":%v", confObj.Port)
var hasInit = false

func initTestConf() {
	if !hasInit {
		initConfig()
		tmpls = initTemplates()
		logToNull()
		hasInit = true
	}
}

func logToNull() {
	hush, err := os.Open("/dev/null")
	if err != nil {
		log.Printf("%v\n", err)
	}
	log.SetOutput(hush)
}
