package main

import (
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

// handles "/"
func indexHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", htmlutf8)
	err := tmpls.ExecuteTemplate(w, "index.html", confObj.Instance)
	if err != nil {
		log.Printf("Error writing to HTTP stream: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

// handles "/api"
func apiBaseHandler(w http.ResponseWriter, r *http.Request) {
	timerfc3339, err := time.Now().MarshalText()
	if err != nil {
		log.Printf("Couldn't format time as RFC3339: %v\n", err)
	}
	etag := fmt.Sprintf("%x", sha256.Sum256(timerfc3339))
	w.Header().Set("ETag", etag)
	w.Header().Set("Content-Type", txtutf8)
	pathdata := []byte("\n\n" + r.URL.Path)
	timerfc3339 = append(timerfc3339, pathdata...)
	n, err := w.Write(timerfc3339)
	if err != nil || n == 0 {
		log.Printf("Error writing to HTTP stream: %v bytes,  %v\n", n, err)
	}
}

// handles "/api/plain"
// maybe add json/xml support later
func apiFormatHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	format := vars["format"]

	w.Header().Set("Content-Type", txtutf8)
	n, err := w.Write([]byte(format + "\n"))
	if err != nil || n == 0 {
		log.Printf("Error writing to HTTP stream: %v bytes,  %v\n", n, err)
	}
}

// handles "/api/plain/(users|mentions|tweets)"
func apiEndpointHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	format := vars["format"]
	endpoint := vars["endpoint"]

	w.Header().Set("Content-Type", htmlutf8)
	n, err := w.Write([]byte(format + "/" + endpoint))
	if err != nil || n == 0 {
		log.Printf("Error writing to HTTP stream: %v bytes,  %v\n", n, err)
	}

}

// handles POST for "/api/plain/users"
func apiEndpointPOSTHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	format := vars["format"]
	endpoint := vars["endpoint"]

	w.Header().Set("Content-Type", htmlutf8)
	n, err := w.Write([]byte(format + "/" + endpoint))
	if err != nil || n == 0 {
		log.Printf("Error writing to HTTP stream: %v bytes,  %v\n", n, err)
	}

}

// handles "/api/plain/tags"
func apiTagsBaseHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	format := vars["format"]

	w.Header().Set("Content-Type", htmlutf8)
	n, err := w.Write([]byte("api/" + format + "/tags"))
	if err != nil || n == 0 {
		log.Printf("Error writing to HTTP stream: %v bytes,  %v\n", n, err)
	}

}

// handles "/api/plain/tags/[a-zA-Z0-9]+"
func apiTagsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	format := vars["format"]
	tags := vars["tags"]

	w.Header().Set("Content-Type", htmlutf8)
	n, err := w.Write([]byte("api/" + format + "/tags/" + tags))
	if err != nil || n == 0 {
		log.Printf("Error writing to HTTP stream: %v bytes,  %v\n", n, err)
	}

}

// Serving the stylesheet virtually because
// files aren't served directly.
func cssHandler(w http.ResponseWriter, _ *http.Request) {
	// read the raw bytes of the stylesheet
	css, err := ioutil.ReadFile("assets/style.css")
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("CSS file does not exist: /css request 404\n")
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		log.Printf("%v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get the mod time for the etag header
	stat, err := os.Stat("assets/style.css")
	if err != nil {
		log.Printf("Couldn't stat CSS file to send ETag header: %v\n", err)
	}

	// Sending the sha256 sum of the modtime in hexadecimal for the ETag header
	etag := fmt.Sprintf("%x", sha256.Sum256([]byte(stat.ModTime().String())))

	w.Header().Set("ETag", "\""+etag+"\"")
	w.Header().Set("Content-Type", cssutf8)
	n, err := w.Write(css)
	if err != nil || n == 0 {
		log.Printf("Error writing to HTTP stream: %v bytes,  %v\n", n, err)
	}
}
