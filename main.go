package main

import (
	"log"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

const getwtxt = "0.1"

func main() {
	log.Printf("getwtxt " + getwtxt + "\n")

	serv := mux.NewRouter()

	serv.HandleFunc("/", indexHandler)
	serv.HandleFunc("/{api:(?:api|api/)}", apiBaseHandler)
	serv.HandleFunc("/api/{format:(?:plain|plain/)}", apiFormatHandler)
	serv.HandleFunc("/api/{format:(?:plain)}/{endpoint:(?:mentions|mentions/|users|users/|tweets|tweets/)}", apiEndpointHandler)
	serv.HandleFunc("/api/{format:(?:plain)}/tags/{tags:[a-zA-Z0-9]+}", apiTagsHandler)
	serv.HandleFunc("/api/{format:(?:plain)}/{tagpathfixer:(?:tags|tags/)}", apiTagsBaseHandler)

	log.Fatalln(http.ListenAndServe(":8080", handlers.CompressHandler(serv)))
}
