package web

import (
	"testing"
)

func TestWebLive(t *testing.T) {
	go Serve(":9001")
	for {}
}
