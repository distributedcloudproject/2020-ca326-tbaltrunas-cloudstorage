package network

import (
	// "cloud/datastore"
	"cloud/utils"
	"net/http"
	"strconv"
	"fmt"
	"encoding/json"
	"os"

	"github.com/gorilla/mux"
	gorillaHandlers "github.com/gorilla/handlers"
)

const (
	apiVersion = "v1"
)

type File struct {
	Key		string   `json:"key"`
}

// ServeWebApp starts a backend web application server with the appropriate routers.
func (c *cloud) ServeWebApp(port int) error {
	address := ":" + strconv.Itoa(port)

	// New gorilla router
	r := mux.NewRouter()

	// Add API version as path prefix
	r = r.PathPrefix(fmt.Sprintf("/api/%s/", apiVersion)).Subrouter()

	// Add public routes.
	// Do not require authentication.
	r.HandleFunc("/ping", c.PingHandler)
	r.HandleFunc("/auth/login", c.AuthLoginHandler).Methods(http.MethodPost)

	// Public route with query string token verification.
	d := r.PathPrefix("/downloadfile").Subrouter()
	d.HandleFunc("/{fileID}", c.GetFileDownload).Methods(http.MethodGet).
														 Queries("token", "")
	d.Use(DownloadTokenMiddleware)

	// Add "secret" routes.
	// Require authentication.
	s := r.PathPrefix("/").Subrouter()
	s.Use(AuthenticationMiddleware)
	s.HandleFunc("/auth/refresh", c.AuthRefreshHandler).Methods(http.MethodGet)
	s.HandleFunc("/netinfo", c.NetworkInfoHandler).Methods(http.MethodGet)
	s.HandleFunc("/files", c.GetFiles).Methods(http.MethodGet)
	s.HandleFunc("/files/{fileID}", c.GetFile).Methods(http.MethodGet).
											   Queries("filter", "contents")
	s.HandleFunc("/files", c.CreateFile).Methods(http.MethodPost)
	s.HandleFunc("/downloadlink/{fileID}", c.GetFileDownloadLink).Methods(http.MethodGet)

	// Add gorilla router as handler for all routes.
	http.Handle("/", r)

	// Apply gorilla middleware handlers.
	// FIXME: use utils.GetLogger() writer, not stdout
	h := gorillaHandlers.LoggingHandler(os.Stdout, r)

	utils.GetLogger().Printf("[DEBUG] Cert and key: %s, %s", os.Getenv("SSL_CRT_FILE"), os.Getenv("SSL_KEY_FILE"))
	utils.GetLogger().Printf("[INFO] Web backend listening on address: %s.", address)
	return http.ListenAndServeTLS(address, os.Getenv("SSL_CRT_FILE"), os.Getenv("SSL_KEY_FILE"), h)
	// TODO: set up server with read and write timeouts.
}

func (c *cloud) NetworkInfoHandler(w http.ResponseWriter, req *http.Request) {
	utils.GetLogger().Println("[INFO] NetworkInfoHandler called.")
	w.WriteHeader(http.StatusOK)
	networkName := c.Network().Name
	w.Write([]byte(fmt.Sprintf(`{"name": "%s"}`, networkName)))
}

// Endpoint
// Method
// Required Body:
// Required Query Params:
// Optional Query Params:
// Response
func (c *cloud) CreateFile(w http.ResponseWriter, req *http.Request) {
	utils.GetLogger().Println("[INFO] CreateFile called.")
	w.WriteHeader(http.StatusOK)

	utils.GetLogger().Printf("[DEBUG] URL: %v", req.URL)
	qs := req.URL.Query()
	utils.GetLogger().Printf("[DEBUG] Querystring parameters: %v", qs)
	names, ok := qs["name"]
	if ok {
		utils.GetLogger().Printf("[DEBUG] Name: %v", names)
	}
	sizes, ok := qs["size"]
	if ok {
		utils.GetLogger().Printf("[DEBUG] Size: %v", sizes)
	}
	var size int
	if len(sizes) != 0 {
		// TODO: check that only 1 param
		size, _ = strconv.Atoi(sizes[0])
	}
	// TODO: validation.
	// Param exists or not -> switch flow.
	// Only 1 value of param.

	// f := datastore.NewFile(file, path, size)
	file, _, _ := req.FormFile("file")
	buffer := make([]byte, size)
	file.Read(buffer)
	utils.GetLogger().Printf("[DEBUG] %v", string(buffer))
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

// TODO: check that query param is filter=content
func (c *cloud) GetFile(w http.ResponseWriter, req *http.Request) {
	utils.GetLogger().Println("[INFO] GetFile called.")
	vars := mux.Vars(req)
	fileID := vars["fileID"]
	utils.GetLogger().Printf("[DEBUG] Got file ID: %s.", fileID)

}

// Approach based on: https://codeburst.io/part-1-jwt-to-authenticate-downloadable-files-at-client-8e0b979c9ac1
func (c *cloud) GetFileDownloadLink(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	fileID := vars["fileID"]

	token, err := GenerateDownloadToken(fileID)
	if err != nil {
		utils.GetLogger().Printf("[ERROR] %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	fileURL := fmt.Sprintf("/downloadfile/%s?token=%s", fileID, token)
	w.Write([]byte(fileURL))
}

func DownloadTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Get token from query string
		tokens := req.URL.Query()["token"]
		if len(tokens) != 1 {
			w.WriteHeader(http.StatusBadRequest)
		}
		token := tokens[0]

		// TODO: when verifying token, check that token fileID is same as passed fileID
		err := ValidateToken(token)
		if err != nil {
			utils.GetLogger().Printf("[ERROR] %v", err)
			if err.Error() == "Signature invalid" {
				w.WriteHeader(http.StatusUnauthorized)
			} else {
				w.WriteHeader(http.StatusBadRequest)
			}
			return
		}

		next.ServeHTTP(w, req)
	})
}

func (c *cloud) GetFileDownload(w http.ResponseWriter, req *http.Request) {
	// token should be verified by middleware
	vars := mux.Vars(req)
	fileID := vars["fileID"]

	// should get file by fileID
	// create a mock file
	var fileName string
	var fileContents []byte
	var fileLength int
	if fileID == "test" {
		fileName = "testresponsefile.txt"
		fileContents = []byte("test response file!!!123")
		fileLength = len(fileContents)
	} else {
		return
	}

	// Set headers for download
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.Itoa(fileLength))
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=%s", fileName))

	// Write out contents.
	n, err := w.Write(fileContents)
	if err != nil {
		utils.GetLogger().Printf("[ERROR] %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	if n != fileLength {
		utils.GetLogger().Printf("[WARN] Written file length does not match: %d (want %d)", n, fileLength)
		w.WriteHeader(http.StatusInternalServerError)
	}

}

// PingHandler for /ping.
func (c *cloud) PingHandler(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "pong"}`))
}
