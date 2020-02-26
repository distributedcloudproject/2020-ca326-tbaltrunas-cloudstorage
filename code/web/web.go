package web

import (
	"net/http"

	"github.com/gorilla/mux"
)

// Web backend, a HTTP API.

type HandlersMap map[string]func(http.ResponseWriter, *http.Request)

func Serve(address string, handlers HandlersMap) error {
	r := mux.NewRouter()

	// Attach handlers.
	for path, handlerFunc := range handlers {
		r.HandleFunc(path, handlerFunc)
	}
	r.Methods("GET", "OPTIONS")

	http.Handle("/", r)

	// Serve content.
	return http.ListenAndServe(address, r)
}

