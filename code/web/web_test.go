package web

import (
	"testing"
	"net/http"
)

func TestWebPing(t *testing.T) {
	handlers := make(HandlersMap)
	handlers["/ping"] = PingHandler
	go Serve(":9001", handlers)
	resp, err := http.Get("<h>																																																																																																																																																																																																																																																																						</h>ttp://localhost:9001/ping")
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
