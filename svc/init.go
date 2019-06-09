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

const getwtxt = "0.3.0"

var (
	flagVersion  *bool   = pflag.BoolP("version", "v", false, "Display version information, then exit.")
	flagHelp     *bool   = pflag.BoolP("help", "h", false, "Display the quick-help screen.")
	flagMan      *bool   = pflag.BoolP("manual", "m", false, "Display the configuration manual.")
	flagConfFile *string = pflag.StringP("config", "c", "", "The name/path of the configuration file you wish to use.")
	flagAssets   *string = pflag.StringP("assets", "a", "", "The location of the getwtxt assets directory")
	flagDBPath   *string = pflag.StringP("db", "d", "", "Path to the getwtxt database")
	flagDBType   *string = pflag.StringP("dbtype", "t", "", "Type of database being used")
)

var confObj = &Configuration{}

// signals to close the log file
var closeLog = make(chan bool, 1)

// used to transmit database pointer after
// initialization
var dbChan = make(chan dbase, 1)

var tmpls *template.Template

var twtxtCache = registry.NewIndex()

var remoteRegistries = &RemoteRegistries{
	Mu:   sync.RWMutex{},
	List: make([]string, 0),
}

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
	go cacheAndPush()
	tmpls = initTemplates()
	staticCache = initAssets()
	watchForInterrupt()
	pingAssets()
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

			log.Printf("\n\nCaught %v. Cleaning up ...\n", sigint)
			confObj.Mu.RLock()
			log.Printf("Closing database connection to %v...\n", confObj.DBPath)

			db := <-dbChan

			switch dbType := db.(type) {
			case *dbLevel:
				errLog("", dbType.db.Close())
			case *dbSqlite:
				errLog("", dbType.db.Close())
			}

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
