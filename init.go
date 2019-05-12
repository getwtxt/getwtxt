package main

import (
	"log"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// command line flags
var (
	port      *int    = pflag.IntP("port", "p", 9001, "getwtxt will serve from this port")
	logfile   *string = pflag.StringP("logfile", "l", "getwtxt.log", "File for logging output")
	twtxtfile *string = pflag.StringP("twtxtfile", "f", "/var/twtxt/twtxt.txt", "Registry file for getwtxt")
)

// config object
var confObj = &configuration{}

// signals to close the log file
var closelog = make(chan bool, 1)

func init() {

	initConfig()

	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	initLogging()
}

func initConfig() {

	viper.SetConfigName("getwtxt")

	viper.AddConfigPath(".")
	viper.AddConfigPath("/usr/local/getwtxt")
	viper.AddConfigPath("/etc")
	viper.AddConfigPath("/usr/local/etc")

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
}

func initLogging() {

	logfile, err := os.OpenFile(confObj.logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		log.Printf("Could not open log file: %v\n", err)
	}

	// Listen for the signal to close the log file
	// in a separate thread
	go func() {
		<-closelog
		log.Printf("Closing log file ...\n")
		err = logfile.Close()
		if err != nil {
			log.Printf("Couldn't close log file: %v\n", err)
		}
	}()

	log.SetOutput(logfile)
}

func rebindConfig() {

	// reassign values to the config object
	confObj.port = viper.GetInt("port")
	confObj.logfile = viper.GetString("logfile")
	confObj.twtxtfile = viper.GetString("twtxtfile")

	// signal to close the log file then wait
	closelog <- true
	time.Sleep(50 * time.Millisecond)

	// reinitialize logging
	initLogging()
}
