package svc // import "github.com/getwtxt/getwtxt/svc"

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/getwtxt/registry"
)

// Requests to apiEndpointPOSTHandler are passed off to this
// function. apiPostUser then fetches the twtxt data, then if
// it's an individual user's file, adds it. If it's registry
// output, it scrapes the users/urls/statuses from the remote
// registry before adding each user to the local cache.
func apiPostUser(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		errHTTP(w, r, fmt.Errorf("error parsing values: %v", err.Error()), http.StatusBadRequest)
		return
	}

	nick := r.FormValue("nickname")
	urls := r.FormValue("url")
	if nick == "" || urls == "" {
		errHTTP(w, r, fmt.Errorf("nickname or URL missing"), http.StatusBadRequest)
		return
	}

	uip := getIPFromCtx(r.Context())

	out, remoteRegistry, err := registry.GetTwtxt(urls, twtxtCache.HTTPClient)
	if err != nil {
		errHTTP(w, r, fmt.Errorf("error fetching twtxt Data: %v", err.Error()), http.StatusBadRequest)
		return
	}

	switch remoteRegistry {
	case true:
		if strings.Contains(urls, confObj.Instance.URL) {
			errHTTP(w, r, fmt.Errorf("can't submit this registry to itself"), http.StatusBadRequest)
			break
		}
		remoteRegistries.List = append(remoteRegistries.List, urls)

		if err := twtxtCache.CrawlRemoteRegistry(urls); err != nil {
			errHTTP(w, r, fmt.Errorf("error crawling remote registry: %v", err.Error()), http.StatusInternalServerError)
		} else {
			log200(r)
		}

	case false:
		statuses, err := registry.ParseUserTwtxt(out, nick, urls)
		errLog("Error Parsing User Data: ", err)

		if err := twtxtCache.AddUser(nick, urls, uip, statuses); err != nil {
			errHTTP(w, r, fmt.Errorf("error adding user to cache: %v", err.Error()), http.StatusBadRequest)
			break
		}

		_, err = w.Write([]byte(fmt.Sprintf("200 OK\n")))
		if err != nil {
			errHTTP(w, r, err, http.StatusInternalServerError)
		} else {
			log200(r)
		}
	}
}
