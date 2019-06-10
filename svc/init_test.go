package svc // import "github.com/getwtxt/getwtxt/svc"

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/getwtxt/registry"
	"github.com/spf13/viper"
)

var (
	testport     string
	initTestOnce sync.Once
	initDBOnce   sync.Once
)

func initTestConf() {
	initTestOnce.Do(func() {
		logToNull()

		testConfig()
		tmpls = initTemplates()
		pingAssets()

		confObj.Mu.RLock()
		defer confObj.Mu.RUnlock()
		testport = fmt.Sprintf(":%v", confObj.Port)
	})
}

func initTestDB() {
	initDBOnce.Do(func() {
		initDatabase()
	})
}

func logToNull() {
	hush, err := os.Open("/dev/null")
	if err != nil {
		log.Printf("%v\n", err)
	}
	log.SetOutput(hush)
}

func testConfig() {

	viper.SetConfigName("getwtxt")
	viper.SetConfigType("yml")
	viper.AddConfigPath("..")

	viper.SetDefault("ListenPort", 9001)
	viper.SetDefault("DatabasePath", "getwtxt.db")
	viper.SetDefault("AssetsDirectory", "assets")
	viper.SetDefault("DatabaseType", "leveldb")
	viper.SetDefault("ReCacheInterval", "1h")
	viper.SetDefault("DatabasePushInterval", "5m")
	viper.SetDefault("Instance.SiteName", "getwtxt")
	viper.SetDefault("Instance.OwnerName", "Anonymous Microblogger")
	viper.SetDefault("Instance.Email", "nobody@knows")
	viper.SetDefault("Instance.URL", "https://twtxt.example.com")
	viper.SetDefault("Instance.Description", "A fast, resilient twtxt registry server written in Go!")

	confObj.Mu.Lock()
	defer confObj.Mu.Unlock()

	confObj.Port = viper.GetInt("ListenPort")
	confObj.AssetsDir = "../" + viper.GetString("AssetsDirectory")

	confObj.DBType = strings.ToLower(viper.GetString("DatabaseType"))
	confObj.DBPath = viper.GetString("DatabasePath")
	log.Printf("Using %v database: %v\n", confObj.DBType, confObj.DBPath)

	confObj.CacheInterval = viper.GetDuration("StatusFetchInterval")
	log.Printf("User status fetch interval: %v\n", confObj.CacheInterval)
	confObj.DBInterval = viper.GetDuration("DatabasePushInterval")
	log.Printf("Database push interval: %v\n", confObj.DBInterval)

	confObj.LastCache = time.Now()
	confObj.LastPush = time.Now()
	confObj.Version = Vers

	confObj.Instance.Vers = Vers
	confObj.Instance.Name = viper.GetString("Instance.SiteName")
	confObj.Instance.URL = viper.GetString("Instance.URL")
	confObj.Instance.Owner = viper.GetString("Instance.OwnerName")
	confObj.Instance.Mail = viper.GetString("Instance.Email")
	confObj.Instance.Desc = viper.GetString("Instance.Description")
}

func mockRegistry() {
	twtxtCache = registry.NewIndex()
	statuses, _, _ := registry.GetTwtxt("https://gbmor.dev/twtxt.txt")
	parsed, _ := registry.ParseUserTwtxt(statuses, "gbmor", "https://gbmor.dev/twtxt.txt")
	_ = twtxtCache.AddUser("gbmor", "https://gbmor.dev/twtxt.txt", "1", net.ParseIP("127.0.0.1"), parsed)
}
