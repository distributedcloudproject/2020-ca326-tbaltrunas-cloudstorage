package network

import (
	"cloud/web"
	"net/http"
	"strconv"
	"fmt"
)

// ListenAndServeHTTP starts a HTTP server and routes requests to the cloud.
func (c *cloud) ListenAndServeHTTP(port int) error {
	address := ":" + strconv.Itoa(port)
	handlers := web.HandlersMap{
		"/netinfo": c.NetworkInfoHandler,
	}
	return web.Serve(address, handlers)
}

func (c *cloud) NetworkInfoHandler(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	networkName := c.Network().Name
	w.Write([]byte(fmt.Sprintf(`{"name": "%s"}`, networkName)))
}
