package web

import (
	"net/http"
)

// Web backend, a HTTP API.

// Serve creates a HTTP server that routes requests through the router.
// FIXME: Is this wrapper really necessary?
func Serve(address string, router http.Handler) error {
	return http.ListenAndServe(address, router)
}

