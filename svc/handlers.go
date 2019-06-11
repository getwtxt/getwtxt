package svc // import "github.com/getwtxt/getwtxt/svc"

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/getwtxt/registry"
	"github.com/gorilla/mux"
)

func getEtag(modtime time.Time) string {
	shabytes, err := modtime.MarshalText()
	errLog("", err)
	return fmt.Sprintf("%x", sha256.Sum256(shabytes))
}

func sendStaticEtag(w http.ResponseWriter, isCSS bool) {
	if isCSS {
		etag := getEtag(staticCache.cssMod)
		w.Header().Set("ETag", "\""+etag+"\"")
		w.Header().Set("Content-Time", cssutf8)
		return
	}
	etag := getEtag(staticCache.indexMod)
	w.Header().Set("ETag", "\""+etag+"\"")
	w.Header().Set("Content-Time", htmlutf8)
}

// handles "/" and "/css"
func staticHandler(w http.ResponseWriter, r *http.Request) {
	pingAssets()

	// Take the hex-encoded sha256 sum of the index template's mod time
	// to send as an ETag. If an error occurred when grabbing the file info,
	// the ETag will be empty.
	staticCache.mu.RLock()
	switch r.URL.Path {
	case "/css":
		sendStaticEtag(w, true)
		_, err := w.Write(staticCache.css)
		if err != nil {
			staticCache.mu.RUnlock()
			errHTTP(w, r, err, http.StatusInternalServerError)
			return
		}
	default:
		sendStaticEtag(w, false)
		_, err := w.Write(staticCache.index)
		if err != nil {
			staticCache.mu.RUnlock()
			errHTTP(w, r, err, http.StatusInternalServerError)
			return
		}
	}
	staticCache.mu.RUnlock()

	log200(r)
}

// handles "/api"
func apiBaseHandler(w http.ResponseWriter, r *http.Request) {
	staticHandler(w, r)
}

// handles "/api/plain"
// maybe add json/xml support later
func apiFormatHandler(w http.ResponseWriter, r *http.Request) {
	staticHandler(w, r)
}

func apiAllTweetsHandler(w http.ResponseWriter, r *http.Request) {
	out, err := twtxtCache.QueryAllStatuses()
	if err != nil {
		errHTTP(w, r, err, http.StatusInternalServerError)
	}

	data := parseQueryOut(out)

	etag := fmt.Sprintf("%x", sha256.Sum256(data))
	w.Header().Set("ETag", etag)
	w.Header().Set("Content-Type", txtutf8)

	_, err = w.Write(data)
	if err != nil {
		errHTTP(w, r, err, http.StatusInternalServerError)
		return
	}

	log200(r)
}

// handles "/api/plain/(users|mentions|tweets)"
func apiEndpointHandler(w http.ResponseWriter, r *http.Request) {
	errLog("Error when parsing query values: ", r.ParseForm())

	if r.FormValue("q") != "" || r.FormValue("url") != "" {
		err := apiEndpointQuery(w, r)
		if err != nil {
			errHTTP(w, r, err, http.StatusInternalServerError)
			return
		}
		log200(r)
		return
	}

	var err error
	page := 1
	pageVal := r.FormValue("page")

	switch pageVal {
	case "":
		break
	default:
		page, err = strconv.Atoi(pageVal)
		if err != nil || page < 1 {
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

	case "/api/plain/tweets":
		out, err = twtxtCache.QueryAllStatuses()
		out = registry.ReduceToPage(page, out)

	default:
		errHTTP(w, r, fmt.Errorf("endpoint not found"), http.StatusNotFound)
		return
	}
	errLog("", err)

	data := parseQueryOut(out)

	etag := fmt.Sprintf("%x", sha256.Sum256(data))
	w.Header().Set("ETag", etag)
	w.Header().Set("Content-Type", txtutf8)

	_, err = w.Write(data)
	if err != nil {
		errHTTP(w, r, err, http.StatusInternalServerError)
	} else {
		log200(r)
	}
}

// handles POST for "/api/plain/users"
func apiEndpointPOSTHandler(w http.ResponseWriter, r *http.Request) {
	apiPostUser(w, r)
}

// handles "/api/plain/tags"
func apiTagsBaseHandler(w http.ResponseWriter, r *http.Request) {
	out, err := twtxtCache.QueryInStatus("#")
	if err != nil {
		errHTTP(w, r, err, http.StatusInternalServerError)
		return
	}

	out = registry.ReduceToPage(1, out)
	data := parseQueryOut(out)

	etag := fmt.Sprintf("%x", sha256.Sum256(data))
	w.Header().Set("ETag", etag)
	w.Header().Set("Content-Type", txtutf8)

	_, err = w.Write(data)
	if err != nil {
		errHTTP(w, r, err, http.StatusInternalServerError)
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
		errHTTP(w, r, err, http.StatusInternalServerError)
		return
	}

	log200(r)
}
