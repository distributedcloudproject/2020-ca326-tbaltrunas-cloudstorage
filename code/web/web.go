package web

import (
	"cloud/utils"
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

	http.Handle("/", r)

	// Serve content.
	utils.GetLogger().Printf("[INFO] Web backend listening on address: %s.", address)
	return http.ListenAndServe(address, r)
}

