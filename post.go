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
		_, _ = w.Write([]byte(fmt.Sprintf("400 Bad Request: %v\n", err)))
		log400(r, err)
		return
	}
	nick := r.FormValue("nickname")
	urls := r.FormValue("url")
	if nick == "" || urls == "" {
		_, _ = w.Write([]byte("400 Bad Request: Nickname or URL Missing\n"))
		log400(r, fmt.Errorf("nickname or URL missing"))
		return
	}

	uip := getIPFromCtx(r.Context())

	out, remoteRegistry, err := registry.GetTwtxt(urls)
	if err != nil {
		_, _ = w.Write([]byte(fmt.Sprintf("400 Bad Request: %v\n", err)))
		log400(r, err)
		return
	}

	if remoteRegistry {
		remoteRegistries.Mu.Lock()
		remoteRegistries.List = append(remoteRegistries.List, urls)
		remoteRegistries.Mu.Unlock()

		err := twtxtCache.ScrapeRemoteRegistry(urls)
		if err != nil {
			_, _ = w.Write([]byte(fmt.Sprintf("400 Bad Request: %v\n", err)))
			log400(r, err)
			return
		}
		log200(r)
		return
	}

	statuses, err := registry.ParseUserTwtxt(out, nick, urls)
	if err != nil {
		_, _ = w.Write([]byte(fmt.Sprintf("400 Bad Request: %v\n", err)))
		log400(r, err)
		return
	}

	err = twtxtCache.AddUser(nick, urls, uip, statuses)
	if err != nil {
		_, _ = w.Write([]byte(fmt.Sprintf("400 Bad Request: %v\n", err)))
		log400(r, err)
		return
	}

	log200(r)
	_, _ = w.Write([]byte(fmt.Sprintf("200 OK\n")))
}
