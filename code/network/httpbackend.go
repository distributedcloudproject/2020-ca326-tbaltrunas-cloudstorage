package network

import (
	"cloud/utils"
	"net/http"
	"strconv"
	"fmt"
	"encoding/json"

	"github.com/gorilla/mux"
)

// ListenAndServeHTTP starts a HTTP server and routes requests to the cloud.
func (c *cloud) ListenAndServeHTTP(port int) error {
	address := ":" + strconv.Itoa(port)

	r := mux.NewRouter()
	r.HandleFunc("/auth", c.WebAuthenticationHandler)
	r.HandleFunc("/netinfo", c.NetworkInfoHandler)
	r.HandleFunc("/files", c.GetFiles).Methods(http.MethodGet)

	utils.GetLogger().Printf("[INFO] HTTP backend listening on address: %s.", address)
	http.Handle("/", r)
	return http.ListenAndServe(address, r)
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

type File struct {
	Key		string   `json:"key"`
}

func (c *cloud) GetFiles(w http.ResponseWriter, req *http.Request) {
	utils.GetLogger().Println("[INFO] GetFiles called.")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	files := []*File{
		&File{
		    Key: "photos/animals/cat in a hat.png",
		//   modified: +Moment().subtract(1, "hours"),
		    // size: 1.5 * 1024 * 1024,
		},
		&File{
		    Key: "photos/animals/kitten_ball.png",
		//   modified: +Moment().subtract(3, "days"),
		    // size: 545 * 1024,
		},
		&File{
		    Key: "photos/animals/elephants.png",
		//   modified: +Moment().subtract(3, "days"),
		    // size: 52 * 1024,
		},
		&File{
		    Key: "photos/funny fall.gif",
		//   modified: +Moment().subtract(2, "months"),
		    // size: 13.2 * 1024 * 1024,
		},
		&File{
		    Key: "photos/holiday.jpg",
		  // modified: +Moment().subtract(25, "days"),
		  //   size: 85 * 1024,
		},
		&File{
		    Key: "documents/letter chunks.doc",
		//   modified: +Moment().subtract(15, "days"),
		    // size: 480 * 1024,
		},
		&File{
		    Key: "documents/export.pdf",
		//   modified: +Moment().subtract(15, "days"),
		    // size: 4.2 * 1024 * 1024,
		},
	}

	data, err := json.Marshal(files)
	if err != nil {
		utils.GetLogger().Printf("[ERROR] Error encoding files as JSON: %v.", err)
		// TODO: respond with status code (internal server error?)
	}
	w.Write(data)
}
