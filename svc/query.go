package svc // import "github.com/getwtxt/getwtxt/svc"

import (
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/getwtxt/registry"
	"github.com/gorilla/mux"
)

func apiErrCheck(err error, r *http.Request) {
	if err != nil {
		uip := getIPFromCtx(r.Context())
		log.Printf("*** %v :: %v %v :: %v\n", uip, r.Method, r.URL, err.Error())
	}
}

func dedupe(list []string) []string {
	out := []string{}
	seen := make(map[string]bool)

	for _, e := range list {
		if !seen[e] {
			out = append(out, e)
			seen[e] = true
		}
	}

	return out
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

// apiEndpointQuery is called via apiEndpointHandler when
// the endpoint is "users" and r.FormValue("q") is not empty.
// It queries the registry cache for users or user URLs
// matching the term supplied via r.FormValue("q")
func apiEndpointQuery(w http.ResponseWriter, r *http.Request) error {
	query := r.FormValue("q")
	urls := r.FormValue("url")
	pageVal := r.FormValue("page")
	var out []string
	var err error

	pageVal = strings.TrimSpace(pageVal)
	page, err := strconv.Atoi(pageVal)
	errLog("", err)

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
		if query != "" && urls != "" {
			out = joinQueryOuts(out2)
		}

	case "mentions":
		if urls == "" {
			return fmt.Errorf("missing URL in mention query")
		}
		urls += ">"
		out, err = twtxtCache.QueryInStatus(urls)
		apiErrCheck(err, r)

	case "tweets":
		out = compositeStatusQuery(query, r)

	default:
		return fmt.Errorf("endpoint query, no cases match")
	}

	out = registry.ReduceToPage(page, out)
	data := parseQueryOut(out)
	etag := fmt.Sprintf("%x", sha256.Sum256(data))

	w.Header().Set("ETag", etag)
	w.Header().Set("Content-Type", txtutf8)
	_, err = w.Write(data)

	return err
}

func joinQueryOuts(data ...[]string) []string {
	single := []string{}
	for _, e := range data {
		single = append(single, e...)
	}
	return dedupe(single)
}

func compositeStatusQuery(query string, r *http.Request) []string {
	var wg sync.WaitGroup
	var out, out2, out3 []string
	var err, err2, err3 error

	wg.Add(3)

	query = strings.ToLower(query)
	go func(query string) {
		out, err = twtxtCache.QueryInStatus(query)
		wg.Done()
	}(query)

	query = strings.Title(query)
	go func(query string) {
		out2, err2 = twtxtCache.QueryInStatus(query)
		wg.Done()
	}(query)

	query = strings.ToUpper(query)
	go func(query string) {
		out3, err3 = twtxtCache.QueryInStatus(query)
		wg.Done()
	}(query)

	wg.Wait()

	apiErrCheck(err, r)
	apiErrCheck(err2, r)
	apiErrCheck(err3, r)

	return joinQueryOuts(out, out2, out3)
}
