package main

import (
	"log"
	"net/http"

	"github.com/gorilla/handlers"
)

func main() {
	log.Printf("getwtxt v0.1\n")

	serv := http.NewServeMux()

	serv.HandleFunc("/api/", apiHandler)

	log.Fatalln(http.ListenAndServe(":8080", handlers.CompressHandler(serv)))
}
