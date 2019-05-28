package main

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// handles "/"
func indexHandler(w http.ResponseWriter, r *http.Request) {

	pingAssets()

	etag := fmt.Sprintf("%x", sha256.Sum256([]byte(staticCache.indexMod.String())))

	// Take the hex-encoded sha256 sum of the index template's mod time
	// to send as an ETag. If an error occurred when grabbing the file info,
	// the ETag will be empty.
	w.Header().Set("ETag", "\""+etag+"\"")
	w.Header().Set("Content-Type", htmlutf8)

	_, err := w.Write(staticCache.index)
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
	switch r.URL.Path {
	case "/api/plain/users":
		out, err = twtxtCache.QueryUser("")

	case "/api/plain/mentions":
		out, err = twtxtCache.QueryInStatus("@<")

	default:
		out, err = twtxtCache.QueryAllStatuses()
	}

	data := parseQueryOut(out)
	if err != nil {
		data = []byte("")
	}

	etag := fmt.Sprintf("%x", sha256.Sum256(data))
	w.Header().Set("ETag", etag)
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

	etag := fmt.Sprintf("%x", sha256.Sum256(data))
	w.Header().Set("ETag", etag)
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

	tags = strings.ToLower(tags)
	out, err := twtxtCache.QueryInStatus("#" + tags)
	if err != nil {
		log500(w, r, err)
		return
	}
	tags = strings.Title(tags)
	out2, err := twtxtCache.QueryInStatus("#" + tags)
	if err != nil {
		log500(w, r, err)
		return
	}
	tags = strings.ToUpper(tags)
	out3, err := twtxtCache.QueryInStatus("#" + tags)
	if err != nil {
		log500(w, r, err)
		return
	}

	out = append(out, out2...)
	out = append(out, out3...)
	out = uniq(out)

	data := parseQueryOut(out)

	etag := fmt.Sprintf("%x", sha256.Sum256(data))
	w.Header().Set("ETag", etag)
	w.Header().Set("Content-Type", txtutf8)

	_, err = w.Write(data)
	if err != nil {
		log500(w, r, err)
		return
	}

	log200(r)
}

// Serving the stylesheet virtually because
// files aren't served directly in getwtxt.
func cssHandler(w http.ResponseWriter, r *http.Request) {

	// Sending the sha256 sum of the modtime in hexadecimal for the ETag header
	etag := fmt.Sprintf("%x", sha256.Sum256([]byte(staticCache.cssMod.String())))

	w.Header().Set("ETag", "\""+etag+"\"")
	w.Header().Set("Content-Type", cssutf8)

	pingAssets()

	n, err := w.Write(staticCache.css)
	if err != nil || n == 0 {
		log500(w, r, err)
		return
	}

	log200(r)
}
