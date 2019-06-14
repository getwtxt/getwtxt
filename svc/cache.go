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
	defer confObj.Mu.RUnlock()
	return template.Must(template.ParseFiles(confObj.AssetsDir + "/tmpl/index.html"))
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

	for _, v := range remoteRegistries.List {
		errLog("Error refreshing local copy of remote registry data: ", twtxtCache.CrawlRemoteRegistry(v))
	}
}

// pingAssets checks if the local static assets
// need to be re-cached. If they do, they are
// pulled back into memory from disk.
func pingAssets() {
	confObj.Mu.RLock()
	defer confObj.Mu.RUnlock()
	staticCache.mu.Lock()
	defer staticCache.mu.Unlock()

	cssStat, err := os.Stat(confObj.AssetsDir + "/style.css")
	errLog("", err)
	indexStat, err := os.Stat(confObj.AssetsDir + "/tmpl/index.html")
	errLog("", err)

	if !staticCache.indexMod.Equal(indexStat.ModTime()) {
		tmpls = initTemplates()

		var b []byte
		buf := bytes.NewBuffer(b)
		errLog("", tmpls.ExecuteTemplate(buf, "index.html", confObj.Instance))

		staticCache.index = buf.Bytes()
		staticCache.indexMod = indexStat.ModTime()
	}

	if !staticCache.cssMod.Equal(cssStat.ModTime()) {
		css, err := ioutil.ReadFile(confObj.AssetsDir + "/style.css")
		errLog("", err)

		staticCache.css = css
		staticCache.cssMod = cssStat.ModTime()
	}
}
