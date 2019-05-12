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

	index := mux.NewRouter()
	api := index.PathPrefix("/api").Subrouter()

	index.HandleFunc("/", indexHandler)
	api.HandleFunc("/", apiBaseHandler)
	api.HandleFunc("/{format:(?:plain)}", apiFormatHandler)
	api.Path("/{format:(?:plain)}/{endpoint:(?:mentions|users|tweets)}").
		Queries("url", "{url}", "q", "{query}", "nickname", "{nickname}").
		HandlerFunc(apiEndpointHandler)
	api.HandleFunc("/{format:(?:plain)}/tags", apiTagsBaseHandler)
	api.HandleFunc("/{format:(?:plain)}/tags/{tags:[a-zA-Z0-9]+}", apiTagsHandler)

	portnum := fmt.Sprintf(":%v", confObj.port)

	server := &http.Server{
		Handler:      handlers.CompressHandler(index),
		Addr:         portnum,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	err := server.ListenAndServe()
	if err != nil {
		log.Printf("%v\n", err)
	}
	closelog <- true
}
