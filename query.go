package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

func apiErrCheck(err error, r *http.Request) {
	if err != nil {
		uip := getIPFromCtx(r.Context())
		log.Printf("*** %v :: %v %v :: %v\n", uip, r.Method, r.URL, err.Error())
	}
}

// Takes the output of queries and formats it for
// an HTTP response. Iterates over the string slice,
// appending each entry to a byte slice, and adding
// newlines where appropriate.
func parseQueryOut(out []string) []byte {
	var data []byte

	for i, e := range out {
		data = append(data, []byte(e)...)

		if !strings.HasSuffix(e, "\n") && i != len(out)-1 {
			data = append(data, byte('\n'))
		}
	}

	return data
}

// Removes duplicate statuses from query output
func uniq(str []string) []string {
	keys := make(map[string]bool)
	out := []string{}
	for _, e := range str {
		if _, ok := keys[e]; !ok {
			keys[e] = true
			out = append(out, e)
		}
	}
	return out
}

// apiUserQuery is called via apiEndpointHandler when
// the endpoint is "users" and r.FormValue("q") is not empty.
// It queries the registry cache for users or user URLs
// matching the term supplied via r.FormValue("q")
func apiEndpointQuery(w http.ResponseWriter, r *http.Request) error {
	query := r.FormValue("q")
	urls := r.FormValue("url")
	var out []string
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
		var out2 []string
		if query != "" {
			out, err = twtxtCache.QueryUser(query)
			apiErrCheck(err, r)
		}
		if urls != "" {
			out2, err = twtxtCache.QueryUser(urls)
			apiErrCheck(err, r)
		}

		out = append(out, out2...)
		out = uniq(out)

	case "mentions":
		if urls == "" {
			return fmt.Errorf("missing URL in mention query")
		}
		urls += ">"
		out, err = twtxtCache.QueryInStatus(urls)
		apiErrCheck(err, r)

	case "tweets":
		query = strings.ToLower(query)
		out, err = twtxtCache.QueryInStatus(query)
		apiErrCheck(err, r)

		query = strings.Title(query)
		out2, err := twtxtCache.QueryInStatus(query)
		apiErrCheck(err, r)

		query = strings.ToUpper(query)
		out3, err := twtxtCache.QueryInStatus(query)
		apiErrCheck(err, r)

		out = append(out, out2...)
		out = append(out, out3...)
		out = uniq(out)

	default:
		return fmt.Errorf("endpoint query, no cases match")
	}

	data := parseQueryOut(out)

	etag := fmt.Sprintf("%x", sha256.Sum256(data))
	w.Header().Set("ETag", etag)
	w.Header().Set("Content-Type", txtutf8)

	_, err = w.Write(data)

	return err
}
