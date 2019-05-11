package main

import (
	"log"
	"net/http"
)

func main() {
	log.Printf("getwtxt v0.1\n")

	http.HandleFunc("/api/", apiHandler)

	log.Fatalln(http.ListenAndServe(":8080", nil))
}
