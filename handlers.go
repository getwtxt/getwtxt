package main

import (
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

// handles "/"
func indexHandler(w http.ResponseWriter, r *http.Request) {

	// Stat the index template to get the mod time
	var etag string
	if indextmpl, err := os.Stat("assets/tmpl/index.html"); err != nil {
		log.Printf("Couldn't stat index template for ETag ... %v\n", err)
	} else {
		etag = fmt.Sprintf("%x", sha256.Sum256([]byte(indextmpl.ModTime().String())))
	}

	// Take the hex-encoded sha256 sum of the index template's mod time
	// to send as an ETag. If an error occured when grabbing the file info,
	// the ETag will be empty.
	w.Header().Set("ETag", "\""+etag+"\"")
	w.Header().Set("Content-Type", htmlutf8)

	// Pass the confObj.Instance data to the template,
	// then send it to the client.
	err := tmpls.ExecuteTemplate(w, "index.html", confObj.Instance)
	if err != nil {
		log500(w, r, err)
		return
	}

	log200(r)
}

// handles "/api"
func apiBaseHandler(w http.ResponseWriter, r *http.Request) {
	indexHandler(w, r)
}

// handles "/api/plain"
// maybe add json/xml support later
func apiFormatHandler(w http.ResponseWriter, r *http.Request) {
	indexHandler(w, r)
}

// handles "/api/plain/(users|mentions|tweets)"
func apiEndpointHandler(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		log500(w, r, err)
		return
	}

	// if there's a query, execute it
	if r.FormValue("q") != "" || r.FormValue("url") != "" {
		err := apiEndpointQuery(w, r)
		if err != nil {
			log500(w, r, err)
			return
		}
		log200(r)
		return
	}

	// if there's no query, return everything in
	// registry for a given endpoint
	var out []string
	if r.URL.Path == "/api/plain/users" {
		out, err = twtxtCache.QueryUser("")
	} else if r.URL.Path == "/api/plain/mentions" {
		out, err = twtxtCache.QueryInStatus(".txt>")
	} else {
		out, err = twtxtCache.QueryAllStatuses()
	}
	if err != nil {
		log500(w, r, err)
		return
	}

	data := parseQueryOut(out)

	w.Header().Set("Content-Type", txtutf8)

	_, err = w.Write(data)
	if err != nil {
		log500(w, r, err)
		return
	}

	log200(r)
}

// handles POST for "/api/plain/users"
func apiEndpointPOSTHandler(w http.ResponseWriter, r *http.Request) {
	apiPostUser(w, r)
}

// handles "/api/plain/tags"
func apiTagsBaseHandler(w http.ResponseWriter, r *http.Request) {

	out, err := twtxtCache.QueryInStatus("#")
	if err != nil {
		log500(w, r, err)
		return
	}

	data := parseQueryOut(out)

	w.Header().Set("Content-Type", txtutf8)

	_, err = w.Write(data)
	if err != nil {
		log500(w, r, err)
		return
	}

	log200(r)
}

// handles "/api/plain/tags/[a-zA-Z0-9]+"
func apiTagsHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	tags := vars["tags"]

	out, err := twtxtCache.QueryInStatus(tags)
	if err != nil {
		log500(w, r, err)
		return
	}

	data := parseQueryOut(out)

	w.Header().Set("Content-Type", txtutf8)
	_, err = w.Write(data)
	if err != nil {
		log500(w, r, err)
		return
	}

	log200(r)
}

// Serving the stylesheet virtually because
// files aren't served directly.
func cssHandler(w http.ResponseWriter, r *http.Request) {

	// read the raw bytes of the stylesheet
	css, err := ioutil.ReadFile("assets/style.css")
	if err != nil {
		if os.IsNotExist(err) {
			log404(w, r, err)
			return
		}
		log500(w, r, err)
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
		log500(w, r, err)
		return
	}

	log200(r)
}
