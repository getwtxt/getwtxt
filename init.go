package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// command line flags
var (
	flagVersion *bool = pflag.BoolP("version", "v", false, "Display version information, then exit")
	flagHelp    *bool = pflag.BoolP("help", "h", false, "")
)

// config object
var confObj = &configuration{}

// signals to close the log file
var closelog = make(chan bool, 1)

func init() {
	checkFlags()
	titleScreen()
	initConfig()
	initLogging()
	watchForInterrupt()
}

func checkFlags() {
	pflag.Parse()
	if *flagVersion {
		titleScreen()
		os.Exit(0)
	}
	if *flagHelp {
		fmt.Printf("\nplaceholder\n")
		fmt.Printf("will add info later\n")
		os.Exit(0)
	}
}

func initConfig() {

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
	confObj.stdoutLogging = viper.GetBool("stdoutLogging")
	confObj.instance.name = viper.GetString("instance.name")
	confObj.instance.url = viper.GetString("instance.url")
	confObj.instance.owner = viper.GetString("instance.owner")
	confObj.instance.mail = viper.GetString("instance.mail")
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
	confObj.stdoutLogging = viper.GetBool("stdoutLogging")
	confObj.instance.name = viper.GetString("instance.name")
	confObj.instance.url = viper.GetString("instance.url")
	confObj.instance.owner = viper.GetString("instance.owner")
	confObj.instance.mail = viper.GetString("instance.mail")

	// reinitialize logging
	initLogging()
}

// Watch for SIGINT aka ^C
// Close the log file then exit
func watchForInterrupt() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		for sigint := range c {
			log.Printf("\n\nCaught %v. Cleaning up ...\n", sigint)

			if !confObj.stdoutLogging {
				// signal to close the log file
				closelog <- true
				time.Sleep(20 * time.Millisecond)
			}

			close(closelog)
			os.Exit(0)
		}
	}()
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
