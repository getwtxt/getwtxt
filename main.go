package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {

	// StrictSlash(true) allows /api and /api/
	// to serve the same content without duplicating
	// handlers/paths
	index := mux.NewRouter().StrictSlash(true)
	api := index.PathPrefix("/api").Subrouter()

	index.Path("/").
		Methods("GET", "HEAD").
		HandlerFunc(indexHandler)

	index.Path("/css").
		Methods("GET", "HEAD").
		HandlerFunc(cssHandler)

	index.Path("/api").
		Methods("GET", "HEAD").
		HandlerFunc(apiBaseHandler)

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
	api.Path("/{format:(?:plain)}/{endpoint:(?:mentions|users|tweets)}").
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

	confObj.Mu.RLock()
	portnum := fmt.Sprintf(":%v", confObj.Port)
	confObj.Mu.RUnlock()

	// handlers.CompressHandler gzips all responses.
	// Write/Read timeouts are self explanatory.
	server := &http.Server{
		Handler:      handlers.CompressHandler(ipMiddleware(index)),
		Addr:         portnum,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Printf("Listening on %v\n", portnum)
	err := server.ListenAndServe()
	if err != nil {
		log.Printf("%v\n", err.Error())
	}

	closeLog <- true
}
