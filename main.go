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
	log.Printf("getwtxt " + getwtxt + "\n")

	// more precise path-based routing
	index := mux.NewRouter()
	api := index.PathPrefix("/api").Subrouter()

	// gorilla/mux makes path validation painless
	index.HandleFunc("/", indexHandler)
	index.HandleFunc("/api", apiBaseHandler)
	api.HandleFunc("/", apiBaseHandler)
	api.HandleFunc("/{format:(?:plain)}", apiFormatHandler)
	api.Path("/{format:(?:plain)}/{endpoint:(?:mentions|users|tweets)}").
		Queries("url", "{url}", "q", "{query}", "nickname", "{nickname}").
		HandlerFunc(apiEndpointHandler)
	api.HandleFunc("/{format:(?:plain)}/tags", apiTagsBaseHandler)
	api.HandleFunc("/{format:(?:plain)}/tags/{tags:[a-zA-Z0-9]+}", apiTagsHandler)

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
