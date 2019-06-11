package svc // import "github.com/getwtxt/getwtxt/svc"

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/spf13/viper"
)

// Configuration values are held in an instance of
// this struct.
type Configuration struct {
	Mu            sync.RWMutex
	IsProxied     bool          `yaml:"BehindProxy"`
	Port          int           `yaml:"ListenPort"`
	LogFile       string        `yaml:"LogFile"`
	DBType        string        `yaml:"DatabaseType"`
	DBPath        string        `yaml:"DatabasePath"`
	AssetsDir     string        `yaml:"-"`
	StdoutLogging bool          `yaml:"StdoutLogging"`
	CacheInterval time.Duration `yaml:"StatusFetchInterval"`
	DBInterval    time.Duration `yaml:"DatabasePushInterval"`
	Instance      `yaml:"Instance"`
}

// Instance refers to meta data about
// this specific instance of getwtxt
type Instance struct {
	Vers  string `yaml:"-"`
	Name  string `yaml:"Instance.SiteName"`
	URL   string `yaml:"Instance.URL"`
	Owner string `yaml:"Instance.OwnerName"`
	Mail  string `yaml:"Instance.Email"`
	Desc  string `yaml:"Instance.Description"`
}

// Called on start-up. Initializes everything
// related to configuration values.
func initConfig() {
	log.Printf("Loading configuration ...\n")

	parseConfigFlag()
	setConfigDefaults()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("%v\n", err.Error())
		log.Printf("Using defaults ...\n")
		bindConfig()
		return
	}

	viper.WatchConfig()
	viper.OnConfigChange(reInit)
	bindConfig()
}

// Registers either stdout or a specified file
// to the default logger.
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
			log.Printf("Closing log file ...\n\n")
			errLog("Could not close log file: ", logfile.Close())
		}(logfile)

		log.SetOutput(logfile)
	}
	confObj.Mu.RUnlock()
}

// Default values should a config file
// not be available.
func setConfigDefaults() {
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
}

// Reads data from the configuration
// flag and acts accordingly.
func parseConfigFlag() {
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
		filename := strings.Split(file, ".")
		viper.SetConfigName(filename[0])
		viper.SetConfigType(filename[1])
		viper.AddConfigPath(path)
	}
}

// Simply goes down the list of fields
// in the confObj instance of &Configuration{},
// assigning values from the config file.
func bindConfig() {
	confObj.Mu.Lock()

	confObj.IsProxied = viper.GetBool("BehindProxy")
	confObj.Port = viper.GetInt("ListenPort")
	confObj.LogFile = viper.GetString("LogFile")
	confObj.DBType = strings.ToLower(viper.GetString("DatabaseType"))
	confObj.DBPath = viper.GetString("DatabasePath")
	confObj.AssetsDir = viper.GetString("AssetsDirectory")
	confObj.StdoutLogging = viper.GetBool("StdoutLogging")
	confObj.CacheInterval = viper.GetDuration("StatusFetchInterval")
	confObj.DBInterval = viper.GetDuration("DatabasePushInterval")

	confObj.Instance.Vers = Vers
	confObj.Instance.Name = viper.GetString("Instance.SiteName")
	confObj.Instance.URL = viper.GetString("Instance.URL")
	confObj.Instance.Owner = viper.GetString("Instance.OwnerName")
	confObj.Instance.Mail = viper.GetString("Instance.Email")
	confObj.Instance.Desc = viper.GetString("Instance.Description")

	if *flagDBType != "" {
		confObj.DBType = *flagDBType
	}
	if *flagDBPath != "" {
		confObj.DBPath = *flagDBPath
	}
	if *flagAssets != "" {
		confObj.AssetsDir = *flagAssets
	}
	if *flagProxied {
		confObj.IsProxied = true
	}

	if confObj.IsProxied {
		log.Printf("Behind reverse proxy, not using host matching\n")
	} else {
		log.Printf("Matching host: %v\n", confObj.Instance.URL)
	}
	if confObj.StdoutLogging {
		log.Printf("Logging to: stdout\n")
	} else {
		log.Printf("Logging to: %v\n", confObj.LogFile)
	}
	log.Printf("Using %v database: %v\n", confObj.DBType, confObj.DBPath)
	log.Printf("Database push interval: %v\n", confObj.DBInterval)
	log.Printf("User status fetch interval: %v\n", confObj.CacheInterval)

	confObj.Mu.Unlock()
}
