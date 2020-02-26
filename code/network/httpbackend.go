package network

import (
	"cloud/web"
	"cloud/utils"
	"net/http"
	"strconv"
	"fmt"
)

// ListenAndServeHTTP starts a HTTP server and routes requests to the cloud.
func (c *cloud) ListenAndServeHTTP(port int) error {
	address := ":" + strconv.Itoa(port)
	handlers := web.HandlersMap{
		"/netinfo": c.NetworkInfoHandler,
		"/auth": c.WebAuthenticationHandler,
	}
	utils.GetLogger().Printf("[INFO] HTTP backend listening on address: %s.", address)
	return web.Serve(address, handlers)
}

func (c *cloud) NetworkInfoHandler(w http.ResponseWriter, req *http.Request) {
	utils.GetLogger().Println("[INFO] NetworkInfoHandler called.")
	w.WriteHeader(http.StatusOK)
	networkName := c.Network().Name
	w.Write([]byte(fmt.Sprintf(`{"name": "%s"}`, networkName)))
}

func (c *cloud) WebAuthenticationHandler(w http.ResponseWriter, req *http.Request) {
	utils.GetLogger().Println("[INFO] WebAuthenticationHandler called.")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write([]byte(`{"access_token": FAKETOKEN"}`))
}
