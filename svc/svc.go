package svc // import "github.com/getwtxt/getwtxt/svc"

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// Start is the initialization function for getwtxt
func Start() {
	before := time.Now()
	initSvc()

	// StrictSlash(true) allows /api and /api/
	// to serve the same content without duplicating
	// handlers/paths
	index := mux.NewRouter().StrictSlash(true)

	confObj.Mu.RLock()
	portnum := fmt.Sprintf(":%v", confObj.Port)
	if !confObj.IsProxied {
		index.Host(confObj.Instance.URL)
	}
	TLS := confObj.TLS.Use
	TLSCert := confObj.TLS.Cert
	TLSKey := confObj.TLS.Key
	confObj.Mu.RUnlock()

	setIndexRouting(index)
	api := index.PathPrefix("/api").Subrouter()
	setEndpointRouting(api)

	server := newServer(portnum, index)
	log.Printf("*** Listening on %v\n", portnum)
	log.Printf("*** getwtxt %v Startup finished at %v, took %v\n\n", Vers, time.Now().Format(time.RFC3339), time.Since(before))
	if TLS {
		errLog("", server.ListenAndServeTLS(TLSCert, TLSKey))
	} else {
		errLog("", server.ListenAndServe())
	}

	closeLog <- true
	killTickers()
	killDB()
	close(dbChan)
	close(closeLog)
}

func newServer(port string, index *mux.Router) *http.Server {
	// handlers.CompressHandler gzips all responses.
	// Write/Read timeouts are self explanatory.
	return &http.Server{
		Handler:      handlers.CompressHandler(ipMiddleware(index)),
		Addr:         port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
}

func setIndexRouting(index *mux.Router) {
	index.Path("/").
		Methods("GET", "HEAD").
		HandlerFunc(staticHandler)

	index.Path("/css").
		Methods("GET", "HEAD").
		HandlerFunc(staticHandler)

	index.Path("/api").
		Methods("GET", "HEAD").
		HandlerFunc(apiBaseHandler)
}

func setEndpointRouting(api *mux.Router) {
	// twtxt will add support for other formats later.
	// Maybe json? Making this future-proof.
	api.Path("/{format:(?:plain)}").
		Methods("GET", "HEAD").
		HandlerFunc(apiFormatHandler)

	// Non-standard API call to list *all* tweets
	// in a single request.
	api.Path("/{format:(?:plain)}/tweets/all").
		Methods("GET", "HEAD").
		HandlerFunc(apiAllTweetsHandler)

	// Specifying the endpoint with and without query information.
	// Will return 404 on empty queries otherwise.
	api.Path("/{format:(?:plain)}/{endpoint:(?:mentions|users|tweets)}").
		Methods("GET", "HEAD").
		HandlerFunc(apiEndpointHandler)

	api.Path("/{format:(?:plain)}/{endpoint:(?:mentions|users|tweets)}").
		Queries("url", "{url}", "q", "{query}", "page", "{[0-9]+}").
		Methods("GET", "HEAD").
		HandlerFunc(apiEndpointHandler)

	// This is for submitting new users. Both query variables must exist
	// in the request for this to match.
	api.Path("/{format:(?:plain)}/{endpoint:users}").
		Queries("url", "{url}", "nickname", "{nickname:[a-zA-Z0-9_-]+}").
		Methods("POST").
		HandlerFunc(apiEndpointPOSTHandler)

	// This is for submitting new users incorrectly
	// and letting the requester know about their error.
	api.Path("/{format:(?:plain)}/{endpoint:users}").
		Queries("url", "{url}").
		Methods("POST").
		HandlerFunc(apiEndpointPOSTHandler)

	// This is also for submitting new users incorrectly
	// and letting the requester know about their error.
	api.Path("/{format:(?:plain)}/{endpoint:users}").
		Queries("nickname", "{nickname:[a-zA-Z0-9_-]+}").
		Methods("POST").
		HandlerFunc(apiEndpointPOSTHandler)

	// Show all observed tags
	api.Path("/{format:(?:plain)}/tags").
		Methods("GET", "HEAD").
		HandlerFunc(apiTagsBaseHandler)

	// Show Nth page of all observed tags
	api.Path("/{format:(?:plain)}/tags").
		Queries("page", "{[0-9]+}").
		Methods("GET", "HEAD").
		HandlerFunc(apiTagsBaseHandler)

	// Requests statuses with a specific tag
	api.Path("/{format:(?:plain)}/tags/{tags:[a-zA-Z0-9_-]+}").
		Methods("GET", "HEAD").
		HandlerFunc(apiTagsHandler)

	// Requests Nth page of statuses with a specific tag
	api.Path("/{format:(?:plain)}/tags/{tags:[a-zA-Z0-9_-]+}").
		Queries("page", "{[0-9]+}").
		Methods("GET", "HEAD").
		HandlerFunc(apiTagsHandler)
}
