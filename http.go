package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"strings"
)

// Attaches a request's IP address to the request's context
func newCtxUserIP(ctx context.Context, r *http.Request) context.Context {

	base := strings.Split(r.RemoteAddr, ":")
	uip := base[0]

	return context.WithValue(ctx, ctxKey, uip)
}

// Retrieves a request's IP address from the request's context
func getIPFromCtx(ctx context.Context) net.IP {

	uip, ok := ctx.Value(ctxKey).(string)
	if !ok {
		log.Printf("Couldn't retrieve IP from request\n")
	}

	return net.ParseIP(uip)
}

// Shim function to modify/pass context value to a handler
func ipMiddleware(hop http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := newCtxUserIP(r.Context(), r)
		hop.ServeHTTP(w, r.WithContext(ctx))
	})
}

// log output for 200s
func log200(r *http.Request) {

	uip := getIPFromCtx(r.Context())
	log.Printf("*** %v :: 200 :: %v %v\n", uip, r.Method, r.URL)
}

// log output for 400s
func log400(r *http.Request, err error) {
	uip := getIPFromCtx(r.Context())
	log.Printf("*** %v :: 400 :: %v %v :: %v\n", uip, r.Method, r.URL, err)
}

// log output for 404s
func log404(w http.ResponseWriter, r *http.Request, err error) {

	uip := getIPFromCtx(r.Context())
	log.Printf("*** %v :: 404 :: %v %v :: %v\n", uip, r.Method, r.URL, err)
	http.Error(w, err.Error(), http.StatusNotFound)
}

// log output for 500s
func log500(w http.ResponseWriter, r *http.Request, err error) {

	uip := getIPFromCtx(r.Context())
	log.Printf("*** %v :: 500 :: %v %v :: %v\n", uip, r.Method, r.URL, err)
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
