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
