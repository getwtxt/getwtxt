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
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// Start is the initialization function for getwtxt
func Start() {
	before := time.Now()
	initSvc()

	// StrictSlash(true) allows /api and /api/
	// to serve the same content without duplicating
	// handlers/paths
	index := mux.NewRouter().StrictSlash(true)
	setIndexRouting(index)
	api := index.PathPrefix("/api").Subrouter()
	setEndpointRouting(api)

	confObj.Mu.RLock()
	portnum := fmt.Sprintf(":%v", confObj.Port)
	if !confObj.IsProxied {
		index.Host(confObj.Instance.URL)
	}
	TLS := confObj.TLS.Use
	TLSCert := confObj.TLS.Cert
	TLSKey := confObj.TLS.Key
	confObj.Mu.RUnlock()

	server := newServer(portnum, index)

	if TLS {
		cert, err := tls.LoadX509KeyPair(TLSCert, TLSKey)
		errFatal("", err)

		cfg := &tls.Config{Certificates: []tls.Certificate{cert}}
		lstnr, err := tls.Listen("tcp", portnum, cfg)
		errFatal("", err)

		server.TLSConfig = cfg
		startAnnounce(portnum, before)
		errLog("", server.ServeTLS(lstnr, "", ""))

	} else {
		startAnnounce(portnum, before)
		errLog("", server.ListenAndServe())
	}

	closeLog <- struct{}{}
	killTickers()
	killDB()
	close(dbChan)
	close(closeLog)
}

func startAnnounce(portnum string, before time.Time) {
	log.Printf("*** Listening on %v\n", portnum)
	log.Printf("*** getwtxt %v Startup finished at %v, took %v\n\n", Vers, time.Now().Format(time.RFC3339), time.Since(before))
}

func newServer(port string, index *mux.Router) *http.Server {
	// handlers.CompressHandler gzips all responses.
	// ipMiddleware passes the request IP along.
	// Write/Read timeouts are self explanatory.
	return &http.Server{
		Handler:      handlers.CompressHandler(ipMiddleware(index)),
		Addr:         port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
}

func setIndexRouting(index *mux.Router) {
	index.Path("/").
		Methods("GET", "HEAD").
		HandlerFunc(staticHandler)
	index.Path("/css").
		Methods("GET", "HEAD").
		HandlerFunc(staticHandler)
	index.Path("/api").
		Methods("GET", "HEAD").
		HandlerFunc(apiBaseHandler)
}

func setEndpointRouting(api *mux.Router) {
	// twtxt will add support for other formats later.
	// Maybe json? Making this future-proof.
	api.Path("/{format:(?:plain)}").
		Methods("GET", "HEAD").
		HandlerFunc(apiFormatHandler)

	// Non-standard API call to list *all* tweets
	// in a single request.
	api.Path("/{format:(?:plain)}/tweets/all").
		Methods("GET", "HEAD").
		HandlerFunc(apiAllTweetsHandler)

	// Specifying the endpoint with and without query information.
	// Will return 404 on empty queries otherwise.
	api.Path("/{format:(?:plain)}/{endpoint:(?:mentions|users|tweets|version)}").
		Methods("GET", "HEAD").
		HandlerFunc(apiEndpointHandler)
	api.Path("/{format:(?:plain)}/{endpoint:(?:mentions|users|tweets)}").
		Queries("url", "{url}", "q", "{query}", "page", "{[0-9]+}").
		Methods("GET", "HEAD").
		HandlerFunc(apiEndpointHandler)

	// This is for submitting new users. Both query variables must exist
	// in the request for this to match.
	api.Path("/{format:(?:plain)}/{endpoint:users}").
		Queries("url", "{url}", "nickname", "{nickname:[a-zA-Z0-9_-]+}").
		Methods("POST").
		HandlerFunc(apiEndpointPOSTHandler)
	// This is for submitting new users incorrectly
	// and letting the requester know about their error.
	api.Path("/{format:(?:plain)}/{endpoint:users}").
		Queries("url", "{url}").
		Methods("POST").
		HandlerFunc(apiEndpointPOSTHandler)
	// This is also for submitting new users incorrectly
	// and letting the requester know about their error.
	api.Path("/{format:(?:plain)}/{endpoint:users}").
		Queries("nickname", "{nickname:[a-zA-Z0-9_-]+}").
		Methods("POST").
		HandlerFunc(apiEndpointPOSTHandler)

	// Show all observed tags
	api.Path("/{format:(?:plain)}/tags").
		Methods("GET", "HEAD").
		HandlerFunc(apiTagsBaseHandler)
	// Show Nth page of all observed tags
	api.Path("/{format:(?:plain)}/tags").
		Queries("page", "{[0-9]+}").
		Methods("GET", "HEAD").
		HandlerFunc(apiTagsBaseHandler)

	// Requests statuses with a specific tag
	api.Path("/{format:(?:plain)}/tags/{tags:[a-zA-Z0-9_-]+}").
		Methods("GET", "HEAD").
		HandlerFunc(apiTagsHandler)
	// Requests Nth page of statuses with a specific tag
	api.Path("/{format:(?:plain)}/tags/{tags:[a-zA-Z0-9_-]+}").
		Queries("page", "{[0-9]+}").
		Methods("GET", "HEAD").
		HandlerFunc(apiTagsHandler)
}
