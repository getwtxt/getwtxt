package main

import (
	"fmt"
	"log"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// command line flags
var (
	port          *int    = pflag.IntP("port", "p", 9001, "getwtxt will serve from this port")
	logfile       *string = pflag.StringP("logfile", "l", "getwtxt.log", "File for logging output")
	twtxtfile     *string = pflag.StringP("twtxtfile", "f", "/var/twtxt/twtxt.txt", "Registry file for getwtxt")
	stdoutLogging *bool   = pflag.BoolP("stdout", "o", true, "Log to stdout rather than to a file")
)

// config object
var confObj = &configuration{}

// signals to close the log file
var closelog = make(chan bool, 1)

func init() {
	titleScreen()
	initConfig()
	initLogging()
}

func initConfig() {

	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	viper.SetConfigName("getwtxt")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/usr/local/getwtxt")
	viper.AddConfigPath("/etc")
	viper.AddConfigPath("/usr/local/etc")

	log.Printf("Loading configuration ...\n")
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Error reading config file: %v\n", err)
		log.Printf("Using defaults ...\n")
	}

	// separate thread to watch for config file changes.
	// will log event then run rebindConfig()
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Printf("Config file change detected. Reloading...\n")
		rebindConfig()
	})

	viper.SetDefault("port", 9001)
	viper.SetDefault("logfile", "getwtxt.log")
	viper.SetDefault("twtxtfile", "/var/twtxt/twtxt.txt")

	confObj.port = viper.GetInt("port")
	confObj.logfile = viper.GetString("logfile")
	confObj.twtxtfile = viper.GetString("twtxtfile")
	confObj.stdoutLogging = viper.GetBool("stdoutLogging")
}

func initLogging() {

	// only open a log file if it's necessary
	if confObj.stdoutLogging {
		log.SetOutput(os.Stdout)

	} else {

		logfile, err := os.OpenFile(confObj.logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			log.Printf("Could not open log file: %v\n", err)
		}

		// Listen for the signal to close the log file
		// in a separate thread. Passing it as an argument
		// to prevent race conditions when the config is
		// reloaded.
		go func(logfile *os.File) {
			<-closelog
			log.Printf("Closing log file ...\n")
			err = logfile.Close()
			if err != nil {
				log.Printf("Couldn't close log file: %v\n", err)
			}
		}(logfile)

		log.SetOutput(logfile)
	}
}

func rebindConfig() {

	// signal to close the log file then wait
	if !confObj.stdoutLogging {
		closelog <- true
	}

	// reassign values to the config object
	confObj.port = viper.GetInt("port")
	confObj.logfile = viper.GetString("logfile")
	confObj.twtxtfile = viper.GetString("twtxtfile")
	confObj.stdoutLogging = viper.GetBool("stdoutLogging")

	// reinitialize logging
	initLogging()
}

func titleScreen() {
	fmt.Printf(`
	
            _            _        _
  __ _  ___| |___      _| |___  _| |_
 / _  |/ _ \ __\ \ /\ / / __\ \/ / __|
| (_| |  __/ |_ \ V  V /| |_ >  <| |_
 \__, |\___|\__| \_/\_/  \__/_/\_\\__|
 |___/
             version ` + getwtxt + `
      github.com/gbmor/getwtxt
                GPL v3	
			 
`)
}
