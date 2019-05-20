package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

func apiErrCheck(err error, r *http.Request) {
	if err != nil {
		uip := getIPFromCtx(r.Context())
		log.Printf("%v :: %v %v :: %v\n", uip, r.Method, r.URL, err)
	}
}

func apiErrCheck500(err error, w http.ResponseWriter, r *http.Request) {
	if err != nil {
		uip := getIPFromCtx(r.Context())
		log.Printf("%v :: %v %v :: %v\n", uip, r.Method, r.URL, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// apiUserQuery is called via apiEndpointHandler when
// the endpoint is "users" and r.FormValue("q") is not empty.
// It queries the registry cache for users or user URLs
// matching the term supplied via r.FormValue("q")
func apiEndpointQuery(w http.ResponseWriter, r *http.Request) {
	query := r.FormValue("q")
	urls := r.FormValue("url")
	var out []string
	var out2 []string
	var err error

	vars := mux.Vars(r)
	endpoint := vars["endpoint"]

	// Handle user URL queries first, then nickname queries.
	// Concatenate both outputs if they're both set.
	// Also handle mention queries and status queries.
	// If we made it this far and 'default' is matched,
	// something went very wrong.
	switch endpoint {
	case "users":
		if urls != "" {
			out2, err = twtxtCache.QueryUser(urls)
			out = append(out, out2...)
			apiErrCheck(err, r)
		}
		if query != "" {
			out2, err = twtxtCache.QueryUser(query)
			out = append(out, out2...)
			apiErrCheck(err, r)
		}

	case "mentions":
		out, err = twtxtCache.QueryInStatus(query)
		apiErrCheck(err, r)

	case "tweets":
		out, err = twtxtCache.QueryInStatus(query)
		apiErrCheck(err, r)

	default:
		http.Error(w, "500", http.StatusInternalServerError)
	}

	// iterate over the output. if there aren't
	// explicit newlines, add them.
	var data []byte
	for _, e := range out {
		data = append(data, []byte(e)...)
		if !strings.HasSuffix(e, "\n") {
			data = append(data, byte('\n'))
		}
	}

	w.Header().Set("Content-Type", txtutf8)
	_, err = w.Write(data)
	apiErrCheck500(err, w, r)
}
