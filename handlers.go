package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"time"
)

func validRequest(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := confObj.validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			log.Printf("Invalid API request: %v", r.URL.Path)
			http.Error(w, fmt.Errorf(r.URL.Path).Error(), http.StatusNotFound)
			return
		}
		fn(w, r, m[2])
	}
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
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
