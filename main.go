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

	// gorilla/mux makes path validation painless
	index.Path("/").
		Methods("GET").
		HandlerFunc(indexHandler)
	index.Path("/api").
		Methods("GET").
		HandlerFunc(apiBaseHandler)
	api.Path("/").
		Methods("GET").
		HandlerFunc(apiBaseHandler)
	api.Path("/{format:(?:plain)}").
		Methods("GET").
		HandlerFunc(apiFormatHandler)
	api.Path("/{format:(?:plain)}/{endpoint:(?:mentions|users|tweets)}").
		Methods("GET").
		Queries("url", "{url}", "q", "{query}").
		HandlerFunc(apiEndpointHandler)
	api.Path("/{format:(?:plain)}/{endpoint:users}").
		Methods("POST").
		Queries("url", "{url}", "nickname", "{nickname}").
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
