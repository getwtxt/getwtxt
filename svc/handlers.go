/*
Copyright (c) 2019 Ben Morrison (gbmor)

This file is part of Getwtxt.

Getwtxt is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

Getwtxt is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with Getwtxt.  If not, see <https://www.gnu.org/licenses/>.
*/

package svc // import "git.sr.ht/~gbmor/getwtxt/svc"

import (
	"errors"
	"fmt"
	"hash/fnv"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"git.sr.ht/~gbmor/getwtxt/registry"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

// Takes the modtime of one of the static files, derives
// an FNV hash from it, then truncates it.
func getEtagFromTime(modtime time.Time) string {
	shabytes, err := modtime.MarshalText()
	errLog("", err)
	bytes := fnv.New32().Sum(shabytes)
	return fmt.Sprintf("%x", bytes[:16])
}

// Takes the body of a response to a request, derives an
// FNV hash from it, then truncates it. Used for non-static
// responses.
func getEtag(data []byte) string {
	bytes := fnv.New32().Sum(data)
	if bytes == nil {
		return ""
	}
	if len(bytes) < 16 {
		return fmt.Sprintf("%x", bytes)
	}
	return fmt.Sprintf("%x", bytes[:16])
}

// Serves index.html and the stylesheet to go with it.
func servStatic(w http.ResponseWriter, isCSS bool) error {
	pingAssets()
	staticCache.mu.RLock()
	defer staticCache.mu.RUnlock()

	var etag string
	var body []byte
	var contentType string

	if isCSS {
		etag = getEtagFromTime(staticCache.cssMod)
		contentType = cssutf8
		body = staticCache.css
	} else {
		etag = getEtagFromTime(staticCache.indexMod)
		contentType = htmlutf8
		body = staticCache.index
	}

	w.Header().Set("ETag", "\""+etag+"\"")
	w.Header().Set("Content-Type", contentType)
	_, err := w.Write(body)
	return err
}

// handles "/" and "/css"
func staticHandler(w http.ResponseWriter, r *http.Request) {
	isCSS := strings.Contains(r.URL.Path, "css")
	if err := servStatic(w, isCSS); err != nil {
		errHTTP(w, r, err, http.StatusInternalServerError)
		return
	}
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

// Serves all tweets without pagination.
func apiAllTweetsHandler(w http.ResponseWriter, r *http.Request) {
	out, err := twtxtCache.QueryAllStatuses()
	if err != nil {
		errHTTP(w, r, err, http.StatusInternalServerError)
	}

	data := parseQueryOut(out)
	etag := getEtag(data)

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
		} else {
			log200(r)
		}
		return
	}

	var err error
	page := 1
	pageVal := r.FormValue("page")

	if pageVal != "" {
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

	case "/api/plain/version":
		etag := getEtag([]byte(Vers))
		w.Header().Set("ETag", "\""+etag+"\"")
		w.Header().Set("Content-Type", txtutf8)
		_, err := w.Write([]byte(strings.TrimSpace("getwtxt " + Vers)))
		if err != nil {
			errHTTP(w, r, err, http.StatusInternalServerError)
		} else {
			log200(r)
		}
		return

	default:
		errHTTP(w, r, fmt.Errorf("endpoint not found"), http.StatusNotFound)
		return
	}
	errLog("", err)

	data := parseQueryOut(out)
	etag := getEtag(data)

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
	etag := getEtag(data)

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
	etag := getEtag(data)

	w.Header().Set("ETag", etag)
	w.Header().Set("Content-Type", txtutf8)

	_, err := w.Write(data)
	if err != nil {
		errHTTP(w, r, err, http.StatusInternalServerError)
		return
	}
	log200(r)
}

func handleUserDelete(w http.ResponseWriter, r *http.Request) {
	pass := r.Header.Get("X-Auth")
	if pass == "" {
		errHTTP(w, r, errors.New("unauthorized"), http.StatusUnauthorized)
		return
	}
	confObj.Mu.RLock()
	adminHash := []byte(confObj.AdminPassHash)
	confObj.Mu.RUnlock()

	if err := bcrypt.CompareHashAndPassword(adminHash, []byte(pass)); err != nil {
		errHTTP(w, r, errors.New("unauthorized"), http.StatusUnauthorized)
		return
	}

	r.ParseForm()
	userURL := strings.TrimSpace(r.Form.Get("url"))
	if userURL == "" {
		errHTTP(w, r, errors.New("bad request"), http.StatusBadRequest)
		return
	}
	if _, err := url.Parse(userURL); err != nil {
		errHTTP(w, r, errors.New("bad request"), http.StatusBadRequest)
		return
	}

	if err := delUser(userURL); err != nil {
		return
	}

	w.WriteHeader(200)
	w.Write([]byte("200 OK\n"))
	log200(r)
}
