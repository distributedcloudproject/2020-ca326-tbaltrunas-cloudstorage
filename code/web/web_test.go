package web

import (
	"testing"
	"net/http"
)

func TestWebPing(t *testing.T) {
	http.HandleFunc("/ping", PingHandler)
	go Serve(":9001", nil)
	resp, err := http.Get("http://localhost:9001/")
	if err != nil {
		t.Error(err)
	}
	t.Log(resp)
}

// PingHandler for /ping.
func PingHandler(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "pong"}`))
}
