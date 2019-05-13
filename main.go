package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

const getwtxt = "0.1"

func main() {

	index := mux.NewRouter().StrictSlash(true)
	api := index.PathPrefix("/api").Subrouter()

	// <3 gorilla/mux
	index.Path("/").
		Methods("GET").
		HandlerFunc(indexHandler)
	index.Path("/api").
		Methods("GET").
		HandlerFunc(apiBaseHandler)
	api.Path("/").
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
	// Validating a URL with regex is unwieldly
	api.Path("/{format:(?:plain)}/{endpoint:(?:mentions|users|tweets)}").
		Queries("url", "{url}", "q", "{query}").
		Methods("GET").
		HandlerFunc(apiEndpointHandler)
	// This is for submitting new users
	api.Path("/{format:(?:plain)}/{endpoint:users}").
		Queries("url", "{url}", "nickname", "{nickname:[a-zA-Z0-9]+}").
		Methods("POST").
		HandlerFunc(apiEndpointPOSTHandler)
	api.Path("/{format:(?:plain)}/tags").
		Methods("GET").
		HandlerFunc(apiTagsBaseHandler)
	api.Path("/{format:(?:plain)}/tags/{tags:[a-zA-Z0-9]+}").
		Methods("GET").
		HandlerFunc(apiTagsHandler)

	// format the port for the http.Server object
	portnum := fmt.Sprintf(":%v", confObj.port)
	// defines options for the http server.
	// handlers.CompressHandler gzips all responses.
	// Write/Read timeouts are self explanatory.
	server := &http.Server{
		Handler:      handlers.CompressHandler(index),
		Addr:         portnum,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Printf("Listening on port %v\n", confObj.port)
	err := server.ListenAndServe()
	if err != nil {
		log.Printf("%v\n", err)
	}

	closelog <- true
}
