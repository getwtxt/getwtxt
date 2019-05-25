package main

import (
	"fmt"
	"log"
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
		log400(w, r, err.Error())
		return
	}

	nick := r.FormValue("nickname")
	urls := r.FormValue("url")
	if nick == "" || urls == "" {
		log400(w, r, "Nickname or URL missing")
		return
	}

	uip := getIPFromCtx(r.Context())

	out, remoteRegistry, err := registry.GetTwtxt(urls)
	if err != nil {
		log400(w, r, err.Error())
		return
	}

	if remoteRegistry {
		remoteRegistries.Mu.Lock()
		remoteRegistries.List = append(remoteRegistries.List, urls)
		remoteRegistries.Mu.Unlock()

		if err := twtxtCache.CrawlRemoteRegistry(urls); err != nil {
			log400(w, r, err.Error())
			return
		}
		log200(r)
		return
	}

	statuses, err := registry.ParseUserTwtxt(out, nick, urls)
	if err != nil {
		log400(w, r, err.Error())
		return
	}

	if err := twtxtCache.AddUser(nick, urls, uip, statuses); err != nil {
		log400(w, r, err.Error())
		return
	}

	log200(r)
	_, err = w.Write([]byte(fmt.Sprintf("200 OK\n")))
	if err != nil {
		log.Printf("%v\n", err)
	}
}
