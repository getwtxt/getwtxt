package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"time"
)

func apiBaseHandler(w http.ResponseWriter, r *http.Request) {
	timerfc3339, err := time.Now().MarshalText()
	if err != nil {
		log.Printf("Couldn't format time as RFC3339: %v\n", err)
	}
	etag := fmt.Sprintf("%x", sha256.Sum256(timerfc3339))
	w.Header().Set("ETag", etag)
	w.Header().Set("Content-Type", textutf8)
	pathdata := []byte("\n\n" + r.URL.Path)
	timerfc3339 = append(timerfc3339, pathdata...)
	n, err := w.Write(timerfc3339)
	if err != nil || n == 0 {
		log.Printf("Error writing to HTTP stream: %v\n", err)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {

}
func apiFormatHandler(w http.ResponseWriter, r *http.Request) {

}
func apiEndpointHandler(w http.ResponseWriter, r *http.Request) {

}
func apiTagsBaseHandler(w http.ResponseWriter, r *http.Request) {

}
func apiTagsHandler(w http.ResponseWriter, r *http.Request) {

}
