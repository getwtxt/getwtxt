package svc // import "github.com/getwtxt/getwtxt/svc"

import (
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/getwtxt/registry"
	"github.com/gorilla/mux"
)

func getEtag(modtime time.Time) string {
	shabytes, err := modtime.MarshalText()
	if err != nil {
		log.Printf("%v\n", err.Error())
	}
	return fmt.Sprintf("%x", sha256.Sum256(shabytes))
}

// handles "/"
func indexHandler(w http.ResponseWriter, r *http.Request) {

	pingAssets()

	// Take the hex-encoded sha256 sum of the index template's mod time
	// to send as an ETag. If an error occurred when grabbing the file info,
	// the ETag will be empty.
	staticCache.mu.RLock()
	etag := getEtag(staticCache.indexMod)
	w.Header().Set("ETag", "\""+etag+"\"")
	w.Header().Set("Content-Type", htmlutf8)

	_, err := w.Write(staticCache.index)
	staticCache.mu.RUnlock()
	if err != nil {
		log500(w, r, err)
		return
	}

	log200(r)
}

// Serving the stylesheet virtually because
// files aren't served directly in getwtxt.
func cssHandler(w http.ResponseWriter, r *http.Request) {

	pingAssets()

	staticCache.mu.RLock()
	etag := getEtag(staticCache.cssMod)
	w.Header().Set("ETag", "\""+etag+"\"")
	w.Header().Set("Content-Type", cssutf8)

	_, err := w.Write(staticCache.css)
	staticCache.mu.RUnlock()
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

func apiAllTweetsHandler(w http.ResponseWriter, r *http.Request) {
	out, err := twtxtCache.QueryAllStatuses()
	if err != nil {
		log500(w, r, err)
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

	page := 1
	pageVal := r.FormValue("page")
	if pageVal != "" {
		page, err = strconv.Atoi(pageVal)
		if err != nil || page == 0 {
			page = 1
		}
	}

	// if there's no query, return everything in
	// registry for a given endpoint
	var out []string
	switch r.URL.Path {
	case "/api/plain/users":
		out, err = twtxtCache.QueryUser("")
		out = registry.ReduceToPage(page, out)

	case "/api/plain/mentions":
		out, err = twtxtCache.QueryInStatus("@<")
		out = registry.ReduceToPage(page, out)

	default:
		out, err = twtxtCache.QueryAllStatuses()
		out = registry.ReduceToPage(page, out)
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

	out = registry.ReduceToPage(1, out)
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

	out := compositeStatusQuery("#"+tags, r)

	out = registry.ReduceToPage(1, out)
	data := parseQueryOut(out)

	etag := fmt.Sprintf("%x", sha256.Sum256(data))
	w.Header().Set("ETag", etag)
	w.Header().Set("Content-Type", txtutf8)

	_, err := w.Write(data)
	if err != nil {
		log500(w, r, err)
		return
	}

	log200(r)
}
