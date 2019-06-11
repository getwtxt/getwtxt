package svc // import "github.com/getwtxt/getwtxt/svc"

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
)

// content-type consts
const txtutf8 = "text/plain; charset=utf-8"
const htmlutf8 = "text/html; charset=utf-8"
const cssutf8 = "text/css; charset=utf-8"

// ipCtxKey is the Context value key for user IP addresses
type ipCtxKey int

const ctxKey ipCtxKey = iota

// Attaches a request's IP address to the request's context.
// If getwtxt is behind a reverse proxy, get the last entry
// in the X-Forwarded-For or X-Real-IP HTTP header as the user IP.
func newCtxUserIP(ctx context.Context, r *http.Request) context.Context {
	base := strings.Split(r.RemoteAddr, ":")
	uip := base[0]

	if _, ok := r.Header["X-Forwarded-For"]; ok {
		proxied := r.Header["X-Forwarded-For"]
		base = strings.Split(proxied[len(proxied)-1], ":")
		uip = base[0]
	}

	xRealIP := http.CanonicalHeaderKey("X-Real-IP")
	if _, ok := r.Header[xRealIP]; ok {
		proxied := r.Header[xRealIP]
		base = strings.Split(proxied[len(proxied)-1], ":")
		uip = base[0]
	}

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

func log200(r *http.Request) {
	useragent := r.Header["User-Agent"]
	uip := getIPFromCtx(r.Context())
	log.Printf("*** %v :: 200 :: %v %v :: %v\n", uip, r.Method, r.URL, useragent)
}

func errHTTP(w http.ResponseWriter, r *http.Request, err error, code int) {
	useragent := r.Header["User-Agent"]
	uip := getIPFromCtx(r.Context())
	log.Printf("*** %v :: %v :: %v %v :: %v :: %v\n", uip, code, r.Method, r.URL, useragent, err.Error())
	http.Error(w, fmt.Sprintf("Error %v: %v", code, err.Error()), code)
}
