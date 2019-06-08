package svc // import "github.com/getwtxt/getwtxt/svc"

import (
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// Configuration object definition
type Configuration struct {
	Mu            sync.RWMutex
	Port          int           `yaml:"ListenPort"`
	LogFile       string        `yaml:"LogFile"`
	DBType        string        `yaml:"DatabaseType"`
	DBPath        string        `yaml:"DatabasePath"`
	AssetsDir     string        `yaml:"-"`
	StdoutLogging bool          `yaml:"StdoutLogging"`
	Version       string        `yaml:"-"`
	CacheInterval time.Duration `yaml:"StatusFetchInterval"`
	DBInterval    time.Duration `yaml:"DatabasePushInterval"`
	LastCache     time.Time     `yaml:"-"`
	LastPush      time.Time     `yaml:"-"`
	Instance      `yaml:"Instance"`
}

// Instance refers to this specific instance of getwtxt
type Instance struct {
	Vers  string `yaml:"-"`
	Name  string `yaml:"Instance.SiteName"`
	URL   string `yaml:"Instance.URL"`
	Owner string `yaml:"Instance.OwnerName"`
	Mail  string `yaml:"Instance.Email"`
	Desc  string `yaml:"Instance.Description"`
}

func initTemplates() *template.Template {
	confObj.Mu.RLock()
	assetsDir := confObj.AssetsDir
	confObj.Mu.RUnlock()

	return template.Must(template.ParseFiles(assetsDir + "/tmpl/index.html"))
}

func initLogging() {

	confObj.Mu.RLock()

	if confObj.StdoutLogging {
		log.SetOutput(os.Stdout)

	} else {

		logfile, err := os.OpenFile(confObj.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		errLog("Could not open log file: ", err)

		// Listen for the signal to close the log file
		// in a separate thread. Passing it as an argument
		// to prevent race conditions when the config is
		// reloaded.
		go func(logfile *os.File) {

			<-closeLog

			log.Printf("Closing log file ...\n")
			errLog("Could not close log file: ", logfile.Close())

		}(logfile)

		log.SetOutput(logfile)
	}

	confObj.Mu.RUnlock()
}

func initConfig() {

	if *flagConfFile == "" {
		viper.SetConfigName("getwtxt")
		viper.SetConfigType("yml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("/usr/local/getwtxt")
		viper.AddConfigPath("/etc")
		viper.AddConfigPath("/usr/local/etc")

	} else {
		path, file := filepath.Split(*flagConfFile)
		if path == "" {
			path = "."
		}
		if file == "" {
			file = *flagConfFile
		}
		filename := strings.Split(file, ".")
		viper.SetConfigName(filename[0])
		viper.SetConfigType(filename[1])
		viper.AddConfigPath(path)
	}

	log.Printf("Loading configuration ...\n")
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("%v\n", err.Error())
		log.Printf("Using defaults ...\n")
	} else {
		viper.WatchConfig()
		viper.OnConfigChange(func(e fsnotify.Event) {
			log.Printf("Config file change detected. Reloading...\n")
			rebindConfig()
		})
	}

	viper.SetDefault("ListenPort", 9001)
	viper.SetDefault("LogFile", "getwtxt.log")
	viper.SetDefault("DatabasePath", "getwtxt.db")
	viper.SetDefault("AssetsDirectory", "assets")
	viper.SetDefault("DatabaseType", "leveldb")
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

	confObj.DBType = strings.ToLower(viper.GetString("DatabaseType"))
	if *flagDBType != "" {
		confObj.DBType = *flagDBType
	}

	confObj.DBPath = viper.GetString("DatabasePath")
	if *flagDBPath != "" {
		confObj.DBPath = *flagDBPath
	}
	log.Printf("Using %v database: %v\n", confObj.DBType, confObj.DBPath)

	confObj.AssetsDir = viper.GetString("AssetsDirectory")
	if *flagAssets != "" {
		confObj.AssetsDir = *flagAssets
	}

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

	confObj.Instance.Vers = getwtxt
	confObj.Instance.Name = viper.GetString("Instance.SiteName")
	confObj.Instance.URL = viper.GetString("Instance.URL")
	confObj.Instance.Owner = viper.GetString("Instance.OwnerName")
	confObj.Instance.Mail = viper.GetString("Instance.Email")
	confObj.Instance.Desc = viper.GetString("Instance.Description")

	confObj.Mu.Unlock()

}

func rebindConfig() {

	confObj.Mu.RLock()
	if !confObj.StdoutLogging {
		closeLog <- true
	}
	confObj.Mu.RUnlock()

	confObj.Mu.Lock()

	confObj.DBType = strings.ToLower(viper.GetString("DatabaseType"))
	if *flagDBType != "" {
		confObj.DBType = *flagDBType
	}

	confObj.LogFile = viper.GetString("LogFile")
	confObj.DBPath = viper.GetString("DatabasePath")
	if *flagDBPath != "" {
		confObj.DBPath = *flagDBPath
	}
	confObj.StdoutLogging = viper.GetBool("StdoutLogging")
	confObj.CacheInterval = viper.GetDuration("StatusFetchInterval")
	confObj.DBInterval = viper.GetDuration("DatabasePushInterval")

	confObj.Instance.Name = viper.GetString("Instance.SiteName")
	confObj.Instance.URL = viper.GetString("Instance.URL")
	confObj.Instance.Owner = viper.GetString("Instance.OwnerName")
	confObj.Instance.Mail = viper.GetString("Instance.Email")
	confObj.Instance.Desc = viper.GetString("Instance.Description")

	confObj.Mu.Unlock()

	initLogging()
}
