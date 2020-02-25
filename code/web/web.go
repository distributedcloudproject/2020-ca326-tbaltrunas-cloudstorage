package web

import (
	"cloud/utils"
	"net/http"

	"github.com/gorilla/mux"
)

// Web backend, a HTTP API.

func Serve(address string) {
	r := mux.NewRouter()

	// Attach handlers
	r.HandleFunc("/ping", PingHandler)

	utils.GetLogger().Printf("[INFO] Web backend listening on address: %s.", address)
	http.Handle("/", r)
	http.ListenAndServe(address, r)
}

// PingHandler for /ping.
func PingHandler(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "pong"}`))
}
