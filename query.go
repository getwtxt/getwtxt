package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

func apiErrCheck(err error, r *http.Request) {
	if err != nil {
		uip := getIPFromCtx(r.Context())
		log.Printf("*** %v :: %v %v :: %v\n", uip, r.Method, r.URL, err)
	}
}

// Takes the output of queries and formats it for
// an HTTP response. Iterates over the string slice,
// appending each entry to a byte slice, and adding
// newlines where appropriate.
func parseQueryOut(out []string) []byte {
	var data []byte

	for _, e := range out {
		data = append(data, []byte(e)...)

		if !strings.HasSuffix(e, "\n") {
			data = append(data, byte('\n'))
		}
	}

	return data
}

// apiUserQuery is called via apiEndpointHandler when
// the endpoint is "users" and r.FormValue("q") is not empty.
// It queries the registry cache for users or user URLs
// matching the term supplied via r.FormValue("q")
func apiEndpointQuery(w http.ResponseWriter, r *http.Request) error {
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
		return fmt.Errorf("endpoint query, no cases match")
	}

	data := parseQueryOut(out)

	w.Header().Set("Content-Type", txtutf8)
	_, err = w.Write(data)
	if err != nil {
		return err
	}

	return nil
}
