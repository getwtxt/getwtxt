package main

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/getwtxt/registry"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/syndtr/goleveldb/leveldb"
)

const getwtxt = "0.1"

// command line flags
var (
	flagVersion *bool = pflag.BoolP("version", "v", false, "Display version information, then exit")
	flagHelp    *bool = pflag.BoolP("help", "h", false, "")
)

// config object
var confObj = &configuration{}

// signals to close the log file
var closelog = make(chan bool, 1)

// used to transmit database pointer after
// initialization
var dbChan = make(chan *leveldb.DB, 1)

// templates
var tmpls *template.Template

// registry index
var twtxtCache = registry.NewIndex()

// remote registry listing
var remote = &RemoteRegistries{}

func init() {
	checkFlags()
	titleScreen()
	initConfig()
	initLogging()
	tmpls = initTemplates()
	watchForInterrupt()
}

func checkFlags() {
	pflag.Parse()
	if *flagVersion {
		titleScreen()
		os.Exit(0)
	}
	if *flagHelp {
		titleScreen()
		helpScreen()
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
	viper.SetDefault("logFile", "getwtxt.log")
	viper.SetDefault("databasePath", "getwtxt.db")
	viper.SetDefault("stdoutLogging", false)
	viper.SetDefault("reCacheInterval", "1h")
	viper.SetDefault("databasePushInterval", "5m")

	updateInterval := viper.GetString("reCacheInterval")
	dur, err := time.ParseDuration(updateInterval)
	if err != nil {
		log.Printf("Unable to parse registry cache update interval. Defaulting to hourly. Msg: %v\n", err)
		dur, _ = time.ParseDuration("1h")
	}

	dbPushInterval := viper.GetString("databasePushInterval")
	dbDur, err := time.ParseDuration(dbPushInterval)
	if err != nil {
		log.Printf("Unable to parse database push interval. Defaulting to every five minutes. Msg: %v\n", err)
		dbDur, _ = time.ParseDuration("5m")
	}

	confObj.mu.Lock()
	confObj.port = viper.GetInt("port")
	confObj.logFile = viper.GetString("logFile")
	confObj.dbPath = viper.GetString("databasePath")
	confObj.stdoutLogging = viper.GetBool("stdoutLogging")
	confObj.cacheInterval = dur
	confObj.dbInterval = dbDur
	confObj.lastCache = time.Now()
	confObj.version = getwtxt
	confObj.Instance.Name = viper.GetString("instance.name")
	confObj.Instance.URL = viper.GetString("instance.url")
	confObj.Instance.Owner = viper.GetString("instance.owner")
	confObj.Instance.Mail = viper.GetString("instance.mail")
	confObj.Instance.Desc = viper.GetString("instance.description")
	confObj.mu.Unlock()
}

func initLogging() {

	// only open a log file if it's necessary
	confObj.mu.RLock()

	if confObj.stdoutLogging {
		log.SetOutput(os.Stdout)

	} else {

		logfile, err := os.OpenFile(confObj.logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
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
	confObj.mu.RUnlock()
}

func rebindConfig() {

	// signal to close the log file then wait
	confObj.mu.RLock()
	if !confObj.stdoutLogging {
		closelog <- true
	}
	confObj.mu.RUnlock()

	// re-parse update interval
	nter := viper.GetString("reCacheInterval")
	dur, err := time.ParseDuration(nter)
	if err != nil {
		log.Printf("Unable to parse update interval. Defaulting to once daily. Msg: %v\n", err)
		dur, _ = time.ParseDuration("1h")
	}

	// re-parse database backup interval
	dbPushInterval := viper.GetString("databasePushInterval")
	dbDur, err := time.ParseDuration(dbPushInterval)
	if err != nil {
		log.Printf("Unable to parse database push interval. Defaulting to every five minutes. Msg: %v\n", err)
		dbDur, _ = time.ParseDuration("5m")
	}

	// reassign values to the config object
	confObj.mu.Lock()
	confObj.port = viper.GetInt("port")
	confObj.logFile = viper.GetString("logFile")
	confObj.stdoutLogging = viper.GetBool("stdoutLogging")
	confObj.dbPath = viper.GetString("databasePath")
	confObj.cacheInterval = dur
	confObj.dbInterval = dbDur
	confObj.Instance.Name = viper.GetString("instance.name")
	confObj.Instance.URL = viper.GetString("instance.url")
	confObj.Instance.Owner = viper.GetString("instance.owner")
	confObj.Instance.Mail = viper.GetString("instance.mail")
	confObj.Instance.Desc = viper.GetString("instance.description")
	confObj.mu.Unlock()

	// reinitialize logging
	initLogging()
}

// Parse the HTML templates
func initTemplates() *template.Template {
	return template.Must(template.ParseFiles("assets/tmpl/index.html"))
}

// Pull DB data into cache, if available
func initDatabase() {
	db, err := leveldb.OpenFile(confObj.dbPath, nil)
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	dbChan <- db
}

// Watch for SIGINT aka ^C
// Close the log file then exit
func watchForInterrupt() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		for sigint := range c {
			log.Printf("\n\nCaught %v. Cleaning up ...\n", sigint)

			confObj.mu.RLock()
			if !confObj.stdoutLogging {
				// signal to close the log file
				closelog <- true
			}
			confObj.mu.RUnlock()

			close(closelog)

			// Let everything catch up
			time.Sleep(30 * time.Millisecond)
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
      github.com/getwtxt/getwtxt
               GPL  v3	
`)
}

func helpScreen() {
	fmt.Printf(`
              Help File

  Sections:
`)
}
