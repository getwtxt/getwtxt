package svc // import "github.com/getwtxt/getwtxt/svc"

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var testport = fmt.Sprintf(":%v", confObj.Port)
var hasInit = false

var initTestOnce sync.Once

func initTestConf() {
	initTestOnce.Do(func() {
		testConfig()
		tmpls = testTemplates()
		logToNull()
	})
}

func logToNull() {
	hush, err := os.Open("/dev/null")
	if err != nil {
		log.Printf("%v\n", err)
	}
	log.SetOutput(hush)
}

func testTemplates() *template.Template {
	return template.Must(template.ParseFiles("../assets/tmpl/index.html"))
}

func testConfig() {

	viper.SetConfigName("getwtxt")
	viper.SetConfigType("yml")
	viper.AddConfigPath("..")

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

	confObj.DBPath = viper.GetString("DatabasePath")
	log.Printf("Using %v database: %v\n", confObj.DBType, confObj.DBPath)

	confObj.AssetsDir = "../" + viper.GetString("AssetsDirectory")

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
