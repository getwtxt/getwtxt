package svc // import "github.com/getwtxt/getwtxt/svc"

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"os"
	"sync"
	"time"
)

// These functions and types pertain to the
// in-memory data being used by the registry
// service, such as:
//  - static assets (index.html, style.css)
//  - the registry itself (users, etc)
//  - list of other registries submitted

// RemoteRegistries holds a list of remote registries to
// periodically scrape for new users. The remote registries
// must have been added via POST like a user.
type RemoteRegistries struct {
	Mu   sync.RWMutex
	List []string
}

// staticAssets holda the rendered landing page
// as a byte slice, its on-disk mod time, the
// assets/style.css file as a byte slice, and
// its on-disk mod time.
type staticAssets struct {
	mu       sync.RWMutex
	index    []byte
	indexMod time.Time
	css      []byte
	cssMod   time.Time
}

// Renders the landing page template using
// the info supplied in the configuration
// file's "Instance" section.
func initTemplates() *template.Template {
	confObj.Mu.RLock()
	assetsDir := confObj.AssetsDir
	confObj.Mu.RUnlock()

	return template.Must(template.ParseFiles(assetsDir + "/tmpl/index.html"))
}

func cacheUpdate() {
	// This clusterfuck of mutex read locks is
	// necessary to avoid deadlock. This mess
	// also avoids a panic that would occur
	// should twtxtCache be written to during
	// this loop.
	twtxtCache.Mu.RLock()
	for k := range twtxtCache.Users {
		twtxtCache.Mu.RUnlock()
		errLog("", twtxtCache.UpdateUser(k))
		twtxtCache.Mu.RLock()
	}
	twtxtCache.Mu.RUnlock()

	remoteRegistries.Mu.RLock()
	for _, v := range remoteRegistries.List {
		errLog("Error refreshing local copy of remote registry data: ", twtxtCache.CrawlRemoteRegistry(v))
	}
	remoteRegistries.Mu.RUnlock()
	confObj.Mu.Lock()
	confObj.LastCache = time.Now()
	confObj.Mu.Unlock()
}

// pingAssets checks if the local static assets
// need to be re-cached. If they do, they are
// pulled back into memory from disk.
func pingAssets() {
	confObj.Mu.RLock()
	assetsDir := confObj.AssetsDir
	confObj.Mu.RUnlock()

	cssStat, err := os.Stat(assetsDir + "/style.css")
	errLog("", err)

	indexStat, err := os.Stat(assetsDir + "/tmpl/index.html")
	errLog("", err)

	staticCache.mu.RLock()
	indexMod := staticCache.indexMod
	cssMod := staticCache.cssMod
	staticCache.mu.RUnlock()

	if !indexMod.Equal(indexStat.ModTime()) {
		tmpls = initTemplates()

		var b []byte
		buf := bytes.NewBuffer(b)

		confObj.Mu.RLock()
		errLog("", tmpls.ExecuteTemplate(buf, "index.html", confObj.Instance))
		confObj.Mu.RUnlock()

		staticCache.mu.Lock()
		staticCache.index = buf.Bytes()
		staticCache.indexMod = indexStat.ModTime()
		staticCache.mu.Unlock()
	}

	if !cssMod.Equal(cssStat.ModTime()) {

		css, err := ioutil.ReadFile(assetsDir + "/style.css")
		errLog("", err)

		staticCache.mu.Lock()
		staticCache.css = css
		staticCache.cssMod = cssStat.ModTime()
		staticCache.mu.Unlock()
	}
}
