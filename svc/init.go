package svc // import "github.com/getwtxt/getwtxt/svc"

import (
	"html/template"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/getwtxt/registry"
	"github.com/spf13/pflag"
)

var (
	// Vers contains the version number set at build time
	Vers         string
	flagVersion  *bool   = pflag.BoolP("version", "v", false, "Display version information, then exit.")
	flagHelp     *bool   = pflag.BoolP("help", "h", false, "Display the quick-help screen.")
	flagMan      *bool   = pflag.BoolP("manual", "m", false, "Display the configuration manual.")
	flagConfFile *string = pflag.StringP("config", "c", "", "The name/path of the configuration file you wish to use.")
	flagAssets   *string = pflag.StringP("assets", "a", "", "The location of the getwtxt assets directory")
	flagDBPath   *string = pflag.StringP("db", "d", "", "Path to the getwtxt database")
	flagDBType   *string = pflag.StringP("dbtype", "t", "", "Type of database being used")
)

// Holds the global configuration
var confObj = &Configuration{}

// Signals to close the log file
var closeLog = make(chan bool, 1)

// Used to transmit database pointer
var dbChan = make(chan dbase, 1)

// Used to transmit the wrapped tickers
// corresponding to the in-memory cache
// or the on-disk database.
var dbTickC = make(chan *tick, 1)
var cTickC = make(chan *tick, 1)

// Used to manage the landing page template
var tmpls *template.Template

// Holds the registry data in-memory
var twtxtCache = registry.NewIndex()

// List of other registries submitted to this registry
var remoteRegistries = &RemoteRegistries{
	Mu:   sync.RWMutex{},
	List: make([]string, 0),
}

// In-memory cache of static assets, specifically
// the parsed landing page and the stylesheet.
var staticCache = &staticAssets{}

func errFatal(context string, err error) {
	if err != nil {
		log.Fatalf(context+"%v\n", err.Error())
	}
}

func errLog(context string, err error) {
	if err != nil {
		log.Printf(context+"%v\n", err.Error())
	}
}

// I'm not using init() because it runs
// even during testing and was causing
// problems.
func initSvc() {
	checkFlags()
	titleScreen()

	initConfig()
	initLogging()
	initDatabase()
	tmpls = initTemplates()
	initPersistence()

	pingAssets()
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
	if *flagMan {
		titleScreen()
		helpScreen()
		manualScreen()
		os.Exit(0)
	}
}

// Watch for SIGINT aka ^C
// Close the log file then exit
func watchForInterrupt() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		for sigint := range c {
			log.Printf("Caught %v. Cleaning up ...\n", sigint)

			killTickers()
			killDB()

			confObj.Mu.RLock()
			log.Printf("Closed database connection to %v\n", confObj.DBPath)
			if !confObj.StdoutLogging {
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
