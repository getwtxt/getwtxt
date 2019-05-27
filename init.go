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
	flagVersion  *bool   = pflag.BoolP("version", "v", false, "Display version information, then exit.")
	flagHelp     *bool   = pflag.BoolP("help", "h", false, "Display the help screen")
	flagConfFile *string = pflag.StringP("config", "c", "getwtxt", "The name of the configuration file without an extension.")
	flagConfType *string = pflag.StringP("type", "t", "yml", "The filetype of the configuration file.")
)

// config object
var confObj = &Configuration{}

// signals to close the log file
var closeLog = make(chan bool, 1)

// used to transmit database pointer after
// initialization
var dbChan = make(chan *leveldb.DB, 1)

// templates
var tmpls *template.Template

// registry index
var twtxtCache = registry.NewIndex()

// remote registry listing
var remoteRegistries = &RemoteRegistries{}

// static assets cache
var staticCache = &struct {
	index    []byte
	indexMod time.Time
	css      []byte
	cssMod   time.Time
}{
	index:    nil,
	indexMod: time.Time{},
	css:      nil,
	cssMod:   time.Time{},
}

func initGetwtxt() {
	checkFlags()
	titleScreen()
	initConfig()
	initLogging()
	tmpls = initTemplates()
	initDatabase()
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

	viper.SetConfigName(*flagConfFile)
	viper.SetConfigType(*flagConfType)
	viper.AddConfigPath(".")
	viper.AddConfigPath("/usr/local/getwtxt")
	viper.AddConfigPath("/etc")
	viper.AddConfigPath("/usr/local/etc")

	log.Printf("Loading configuration ...\n")
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("%v\n", err)
		log.Printf("Using defaults ...\n")
	} else {
		// separate thread to watch for config file changes.
		// will log event then run rebindConfig()
		viper.WatchConfig()
		viper.OnConfigChange(func(e fsnotify.Event) {
			log.Printf("Config file change detected. Reloading...\n")
			rebindConfig()
		})
	}

	viper.SetDefault("ListenPort", 9001)
	viper.SetDefault("LogFile", "getwtxt.log")
	viper.SetDefault("DatabasePath", "getwtxt.db")
	viper.SetDefault("StdoutLogging", false)
	viper.SetDefault("ReCacheInterval", "1h")
	viper.SetDefault("DatabasePushInterval", "5m")

	viper.SetDefault("Instance.SiteName", "getwtxt")
	viper.SetDefault("Instance.OwnerName", "Anonymous Microblogger")
	viper.SetDefault("Instance.Email", "nobody@knows")
	viper.SetDefault("Instance.URL", "https://twtxt.example.com")
	viper.SetDefault("Instance.Description", "A fast, resilient twtxt registry server written in Go!")

	confObj.Mu.Lock()

	confObj.Port = viper.GetInt("ListenPort")
	confObj.LogFile = viper.GetString("LogFile")

	confObj.DBPath = viper.GetString("DatabasePath")
	log.Printf("Using database: %v\n", confObj.DBPath)

	confObj.StdoutLogging = viper.GetBool("StdoutLogging")
	if confObj.StdoutLogging {
		log.Printf("Logging to stdout\n")
	} else {
		log.Printf("Logging to %v\n", confObj.LogFile)
	}

	confObj.CacheInterval = viper.GetDuration("StatusFetchInterval")
	log.Printf("User status fetch interval: %v\n", confObj.CacheInterval)

	confObj.DBInterval = viper.GetDuration("DatabasePushInterval")
	log.Printf("Database push interval: %v\n", confObj.DBInterval)

	confObj.LastCache = time.Now()
	confObj.LastPush = time.Now()
	confObj.Version = getwtxt

	confObj.Instance.Name = viper.GetString("Instance.SiteName")
	confObj.Instance.URL = viper.GetString("Instance.URL")
	confObj.Instance.Owner = viper.GetString("Instance.OwnerName")
	confObj.Instance.Mail = viper.GetString("Instance.Email")
	confObj.Instance.Desc = viper.GetString("Instance.Description")

	confObj.Mu.Unlock()

}

func initLogging() {

	// only open a log file if it's necessary
	confObj.Mu.RLock()

	if confObj.StdoutLogging {
		log.SetOutput(os.Stdout)

	} else {

		logfile, err := os.OpenFile(confObj.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			log.Printf("Could not open log file: %v\n", err)
		}

		// Listen for the signal to close the log file
		// in a separate thread. Passing it as an argument
		// to prevent race conditions when the config is
		// reloaded.
		go func(logfile *os.File) {

			<-closeLog
			log.Printf("Closing log file ...\n")

			err = logfile.Close()
			if err != nil {
				log.Printf("Couldn't close log file: %v\n", err)
			}
		}(logfile)

		log.SetOutput(logfile)
	}
	confObj.Mu.RUnlock()
}

func rebindConfig() {

	// signal to close the log file then wait
	confObj.Mu.RLock()
	if !confObj.StdoutLogging {
		closeLog <- true
	}
	confObj.Mu.RUnlock()

	// reassign values to the config object
	confObj.Mu.Lock()

	confObj.LogFile = viper.GetString("LogFile")
	confObj.DBPath = viper.GetString("DatabasePath")
	confObj.StdoutLogging = viper.GetBool("StdoutLogging")
	confObj.CacheInterval = viper.GetDuration("StatusFetchInterval")
	confObj.DBInterval = viper.GetDuration("DatabasePushInterval")

	confObj.Instance.Name = viper.GetString("Instance.SiteName")
	confObj.Instance.URL = viper.GetString("Instance.URL")
	confObj.Instance.Owner = viper.GetString("Instance.OwnerName")
	confObj.Instance.Mail = viper.GetString("Instance.Email")
	confObj.Instance.Desc = viper.GetString("Instance.Description")

	confObj.Mu.Unlock()

	// reinitialize logging
	initLogging()
}

// Parse the HTML templates
func initTemplates() *template.Template {
	return template.Must(template.ParseFiles("assets/tmpl/index.html"))
}

// Pull DB data into cache, if available.
func initDatabase() {
	confObj.Mu.RLock()
	db, err := leveldb.OpenFile(confObj.DBPath, nil)
	confObj.Mu.RUnlock()
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	// Send the database reference into
	// the aether.
	dbChan <- db

	pullDatabase()
	go cacheAndPush()
}

// Watch for SIGINT aka ^C
// Close the log file then exit
func watchForInterrupt() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		for sigint := range c {

			log.Printf("\n\nCaught %v. Cleaning up ...\n", sigint)
			confObj.Mu.RLock()

			// Close the database cleanly
			log.Printf("Closing database connection to %v...\n", confObj.DBPath)
			db := <-dbChan
			if err := db.Close(); err != nil {
				log.Printf("%v\n", err)
			}

			if !confObj.StdoutLogging {
				// signal to close the log file
				closeLog <- true
			}

			confObj.Mu.RUnlock()
			close(dbChan)
			close(closeLog)

			// Let everything catch up
			time.Sleep(100 * time.Millisecond)
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
              getwtxt Help

Command line options:
    -h               Print this help screen.
    -v               Print the version information and quit.
    -c [--config]    Name of an alternate configuration file
                       to use. Do not include the file extention,
                       such as ".yml". Must be in the expected
                       locations.
    -t [--type]      The file type / extension of the config file.
                       json, yml, etc.


`)
}
