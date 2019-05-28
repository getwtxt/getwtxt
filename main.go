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
	initGetwtxt()

	// StrictSlash(true) allows /api and /api/
	// to serve the same content without duplicating
	// handlers/paths
	index := mux.NewRouter().StrictSlash(true)
	api := index.PathPrefix("/api").Subrouter()

	index.Path("/").
		Methods("GET").
		HandlerFunc(indexHandler)

	index.Path("/css").
		Methods("GET").
		HandlerFunc(cssHandler)

	index.Path("/api").
		Methods("GET").
		HandlerFunc(apiBaseHandler)

	// twtxt will add support for other formats later.
	// Maybe json? Making this future-proof.
	api.Path("/{format:(?:plain)}").
		Methods("GET").
		HandlerFunc(apiFormatHandler)

	// Specifying the endpoint with and without query information.
	// Will return 404 on empty queries otherwise.
	api.Path("/{format:(?:plain)}/{endpoint:(?:mentions|users|tweets)}").
		Methods("GET").
		HandlerFunc(apiEndpointHandler)

	// Using stdlib net/url to validate the input URLs rather than regex.
	// Validating a URL with regex is unwieldy
	api.Path("/{format:(?:plain)}/{endpoint:(?:mentions|users|tweets)}").
		Queries("url", "{url}", "q", "{query}").
		Methods("GET").
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

	// This is for submitting new users incorrectly
	// and letting the requester know about their error.
	api.Path("/{format:(?:plain)}/{endpoint:users}").
		Queries("nickname", "{nickname:[a-zA-Z0-9_-]+}").
		Methods("POST").
		HandlerFunc(apiEndpointPOSTHandler)

	// Show all observed tags
	api.Path("/{format:(?:plain)}/tags").
		Methods("GET").
		HandlerFunc(apiTagsBaseHandler)

	// Requests statuses with a specific tag
	api.Path("/{format:(?:plain)}/tags/{tags:[a-zA-Z0-9_-]+}").
		Methods("GET").
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
		log.Printf("%v\n", err)
	}

	closeLog <- true
}
