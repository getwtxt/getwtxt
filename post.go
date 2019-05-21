package main

import (
	"fmt"
	"net/http"

	"github.com/getwtxt/registry"
)

// Requests to apiEndpointPOSTHandler are passed off to this
// function. apiPostUser then fetches the twtxt data, then if
// it's an individual user's file, adds it. If it's registry
// output, it scrapes the users/urls/statuses from the remote
// registry before adding each user to the local cache.
func apiPostUser(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log400(w, r, err)
		return
	}
	nick := r.FormValue("nickname")
	urls := r.FormValue("url")
	if nick == "" || urls == "" {
		log400(w, r, fmt.Errorf("nickname or URL missing"))
		return
	}

	uip := getIPFromCtx(r.Context())

	out, remoteRegistry, err := registry.GetTwtxt(urls)
	if err != nil {
		log400(w, r, err)
		return
	}

	if remoteRegistry {
		remote.Mu.Lock()
		remote.List = append(remote.List, urls)
		remote.Mu.Unlock()

		err := twtxtCache.ScrapeRemoteRegistry(urls)
		if err != nil {
			log400(w, r, err)
			return
		}
		log200(r)
		return
	}

	statuses, err := registry.ParseUserTwtxt(out)
	if err != nil {
		log400(w, r, err)
		return
	}

	err = twtxtCache.AddUser(nick, urls, uip, statuses)
	if err != nil {
		log400(w, r, err)
		return
	}

	log200(r)
}
