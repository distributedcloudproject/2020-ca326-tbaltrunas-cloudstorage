package network

import (
	"cloud/datastore"
	"cloud/utils"
	"net/http"
	"strconv"
	"fmt"
	"encoding/json"
	"os"
	"io/ioutil"
	"io"

	"github.com/gorilla/mux"
	gorillaHandlers "github.com/gorilla/handlers"
)

const (
	apiVersion = "v1"
)

type webapp struct {
	cloud Cloud
}

type WebApp interface {
	Serve(port int) error
}

type File struct {
	Key		string   `json:"key"`
}

func NewWebApp(c Cloud) WebApp {
	 webapp := webapp{
		cloud: c,
	}
	return &webapp
}

// ServeWebApp starts a backend web application server with the appropriate routers.
func (wapp *webapp) Serve(port int) error {
	address := ":" + strconv.Itoa(port)

	// New gorilla router
	r := mux.NewRouter()

	// Add API version as path prefix
	r = r.PathPrefix(fmt.Sprintf("/api/%s/", apiVersion)).Subrouter()

	// Add public routes.
	// Do not require authentication.
	r.HandleFunc("/ping", wapp.WebPingHandler)
	r.HandleFunc("/auth/login", wapp.WebAuthLoginHandler).Methods(http.MethodPost)

	// Public route with query string token verification.
	d := r.PathPrefix("/downloadfile").Subrouter()
	d.Use(DownloadTokenMiddleware)
	d.HandleFunc("/{fileID}", wapp.WebGetFileDownload).Methods(http.MethodGet).
														 Queries("token", "")

	// Add "secret" routes.
	// Require authentication.
	s := r.PathPrefix("/").Subrouter()
	s.Use(AuthenticationMiddleware)
	s.HandleFunc("/auth/refresh", wapp.WebAuthRefreshHandler).Methods(http.MethodGet)
	s.HandleFunc("/netinfo", wapp.WebNetworkInfoHandler).Methods(http.MethodGet)
	s.HandleFunc("/files", wapp.WebGetFiles).Methods(http.MethodGet)
	s.HandleFunc("/files/{fileID}", wapp.WebGetFile).Methods(http.MethodGet).
											   Queries("filter", "contents")

	s.HandleFunc("/files", wapp.CreateFile).Methods(http.MethodPost)

	s.HandleFunc("/downloadlink/{fileID}", wapp.WebGetFileDownloadLink).Methods(http.MethodGet)

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

// FIXME: Instead of having methods of cloud, have a new struct WebApp with cloud as an attribute
// and these methods

func (wapp *webapp) WebNetworkInfoHandler(w http.ResponseWriter, req *http.Request) {
	utils.GetLogger().Println("[INFO] NetworkInfoHandler called.")
	w.WriteHeader(http.StatusOK)
	networkName := wapp.cloud.Network().Name
	w.Write([]byte(fmt.Sprintf(`{"name": "%s"}`, networkName)))
}

// CreateFile API method creates a new file on the cloud.
// Endpoint: /files
// Method: POST.
// Headers: Authorization.
// Query parameters:
// - name=str, the path of the file on the cloud (also the file's key).
// - size=int, the expected size of the file's contents.
// - type=str (optional), the file type (extension).
// - lastModified=date (optional), the date the file was last modified or uploaded.
// Body:
// - File contents as POST body, encoded using post data.
// Response:
// - 200 if file is stored on the cloud successfully.
func (wapp *webapp) CreateFile(w http.ResponseWriter, req *http.Request) {
	utils.GetLogger().Println("[INFO] CreateFile called.")

	utils.GetLogger().Printf("[DEBUG] URL: %v", req.URL)
	qs := req.URL.Query()
	utils.GetLogger().Printf("[DEBUG] Querystring parameters: %v", qs)

	path, err := GetQueryParam(req.URL, "name")
	if err != nil {
		utils.GetLogger().Printf("[ERROR] %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	sizeStr, err := GetQueryParam(req.URL, "size")
	if err != nil {
		utils.GetLogger().Printf("[ERROR] %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		utils.GetLogger().Printf("[ERROR] %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// fType := GetQueryParam(req.URL, "type")
	// lastModified := GetQueryParam(req.URL, "lastModified")

	multipartFileReader, _, _ := req.FormFile("file")
	defer multipartFileReader.Close()

	// Create File data structure
	file, err := datastore.NewFile(multipartFileReader, path, wapp.cloud.Config().FileChunkSize)
	if err != nil {
		utils.GetLogger().Printf("[ERROR] %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	utils.GetLogger().Printf("[DEBUG] Created file: %v", file)
	// Verify size
	if int(file.Size) != size {
		utils.GetLogger().Printf("[ERROR] File sizes do not match: %v (want %v)", file.Size, size)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Temporarily save the file locally.
	tmpFile, err := ioutil.TempFile("", "cloud_web_file_*")
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()
	if err != nil {
		utils.GetLogger().Printf("[ERROR] %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Write out file contents.
	var bufferSize int // pick the smaller buffer size
	if size < wapp.cloud.Config().FileChunkSize {
		bufferSize = size
	} else {
		bufferSize = wapp.cloud.Config().FileChunkSize
	}
	utils.GetLogger().Printf("[DEBUG] Buffer size: %v", bufferSize)
	buffer := make([]byte, bufferSize)
	written := 0
	for written < size {
		numRead, err := multipartFileReader.Read(buffer)
		if err != nil && err != io.EOF {
			utils.GetLogger().Printf("[ERROR] %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		numWritten, err := tmpFile.Write(buffer[:numRead])
		if err != nil {
			utils.GetLogger().Printf("[ERROR] %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		written += numWritten
	}
	localPath := tmpFile.Name();
	utils.GetLogger().Printf("[DEBUG] Saved contents to: %v.", localPath)

	// Finally add the file.
	wapp.cloud.AddFile(file, path, localPath)

	// Send back a response.
	w.WriteHeader(http.StatusOK)
}

func (wapp *webapp) WebGetFiles(w http.ResponseWriter, req *http.Request) {
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
func (wapp *webapp) WebGetFile(w http.ResponseWriter, req *http.Request) {
	utils.GetLogger().Println("[INFO] GetFile called.")
	vars := mux.Vars(req)
	fileID := vars["fileID"]
	utils.GetLogger().Printf("[DEBUG] Got file ID: %s.", fileID)

}

// Approach based on: https://codeburst.io/part-1-jwt-to-authenticate-downloadable-files-at-client-8e0b979c9ac1
func (wapp *webapp) WebGetFileDownloadLink(w http.ResponseWriter, req *http.Request) {
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
		err := ValidateDownloadToken(token)
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

func (wapp *webapp) WebGetFileDownload(w http.ResponseWriter, req *http.Request) {
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
func (wapp *webapp) WebPingHandler(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "pong"}`))
}
