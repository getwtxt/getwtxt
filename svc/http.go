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

// Appends a 200 OK to the request log
func log200(r *http.Request) {
	useragent := r.Header["User-Agent"]
	uip := getIPFromCtx(r.Context())
	reqLog.Printf("*** %v :: 200 :: %v %v :: %v\n", uip, r.Method, r.URL, useragent)
}

// Appends a request of a given status code to the request
// log. Intended for errors.
func errHTTP(w http.ResponseWriter, r *http.Request, err error, code int) {
	useragent := r.Header["User-Agent"]
	uip := getIPFromCtx(r.Context())
	reqLog.Printf("*** %v :: %v :: %v %v :: %v :: %v\n", uip, code, r.Method, r.URL, useragent, err.Error())
	http.Error(w, fmt.Sprintf("Error %v: %v", code, err.Error()), code)
}
