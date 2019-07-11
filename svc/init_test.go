/*
Copyright (c) 2019 Ben Morrison (gbmor)

This file is part of Getwtxt.

Getwtxt is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

Getwtxt is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with Getwtxt.  If not, see <https://www.gnu.org/licenses/>.
*/

package svc // import "github.com/getwtxt/getwtxt/svc"

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"testing"

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
	reqLog = log.New(hush, "", log.LstdFlags)
}

func testConfig() {
	viper.SetConfigName("getwtxt")
	viper.SetConfigType("yml")
	viper.AddConfigPath("../")

	viper.SetDefault("BehindProxy", true)
	viper.SetDefault("UseTLS", false)
	viper.SetDefault("TLSCert", "/etc/ssl/getwtxt.pem")
	viper.SetDefault("TLSKey", "/etc/ssl/private/getwtxt.pem")
	viper.SetDefault("ListenPort", 9001)
	viper.SetDefault("DatabasePath", "getwtxt.db")
	viper.SetDefault("AssetsDirectory", "assets")
	viper.SetDefault("DatabaseType", "leveldb")
	viper.SetDefault("ReCacheInterval", "9m")
	viper.SetDefault("DatabasePushInterval", "4m")
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
	confObj.CacheInterval = viper.GetDuration("StatusFetchInterval")
	confObj.DBInterval = viper.GetDuration("DatabasePushInterval")

	confObj.Instance.Vers = Vers
	confObj.Instance.Name = viper.GetString("Instance.SiteName")
	confObj.Instance.URL = viper.GetString("Instance.URL")
	confObj.Instance.Owner = viper.GetString("Instance.OwnerName")
	confObj.Instance.Mail = viper.GetString("Instance.Email")
	confObj.Instance.Desc = viper.GetString("Instance.Description")
}

// Creates a fresh mock registry, with a single
// user and their statuses, for testing.
func mockRegistry() {
	twtxtCache = registry.New(nil)
	statuses, _, _ := registry.GetTwtxt("https://gbmor.dev/twtxt.txt", nil)
	parsed, _ := registry.ParseUserTwtxt(statuses, "gbmor", "https://gbmor.dev/twtxt.txt")
	_ = twtxtCache.AddUser("gbmor", "https://gbmor.dev/twtxt.txt", net.ParseIP("127.0.0.1"), parsed)
}

// Empties the mock registry's user of statuses
// for functions that test status modifications
func killStatuses() {
	twtxtCache.Mu.Lock()
	user := twtxtCache.Users["https://gbmor.dev/twtxt.txt"]
	user.Mu.Lock()

	user.Status = registry.NewTimeMap()
	user.LastModified = "0"
	twtxtCache.Users["https://gbmor.dev/twtxt.txt"] = user

	user.Mu.Unlock()
	twtxtCache.Mu.Unlock()
}

func Test_errLog(t *testing.T) {
	t.Run("Log to Buffer", func(t *testing.T) {
		b := []byte{}
		buf := bytes.NewBuffer(b)
		log.SetOutput(buf)
		err := fmt.Errorf("test error")
		errLog("", err)
		if !strings.Contains(buf.String(), "test error") {
			t.Errorf("Output Incorrect: %#v\n", buf.String())
		}
	})
}
