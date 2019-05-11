package main

import (
	"log"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {
	log.Printf("getwtxt v0.1\n")

	serv := mux.NewRouter()

	serv.HandleFunc("/", indexHandler)
	serv.HandleFunc("/api/", apiBaseHandler)
	serv.HandleFunc("/api/{format:plain}", apiFormatHandler)
	serv.HandleFunc("/api/{format:plain}/{endpoint:mentions|users|tweets}", apiEndpointHandler)
	serv.HandleFunc("/api/{format:plain}/tags/{tags:[a-zA-Z0-9]+}", apiTagsHandler)
	serv.HandleFunc("/api/{format:plain}/tags", apiTagsBaseHandler)

	log.Fatalln(http.ListenAndServe(":8080", handlers.CompressHandler(serv)))
}
