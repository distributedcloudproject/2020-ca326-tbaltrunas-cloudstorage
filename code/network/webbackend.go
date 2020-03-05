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

type WebFile struct {
	Key	string `json:"key"`
	Size int `json:"size"`
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
	d.HandleFunc("/{fileKey}", wapp.WebGetFileDownload).Methods(http.MethodGet).
														Queries("token", "")

	// Add "secret" routes.
	// Require authentication.
	s := r.PathPrefix("/").Subrouter()
	s.Use(AuthenticationMiddleware)
	s.HandleFunc("/auth/refresh", wapp.WebAuthRefreshHandler).Methods(http.MethodGet)
	s.HandleFunc("/netinfo", wapp.WebNetworkInfoHandler).Methods(http.MethodGet)

	s.HandleFunc("/files", wapp.CreateFile).Methods(http.MethodPost)
	s.HandleFunc("/files", wapp.ReadFiles).Methods(http.MethodGet)
	s.HandleFunc("/files/{fileKey}", wapp.WebGetFile).Methods(http.MethodGet).
											   Queries("filter", "contents")

	s.HandleFunc("/files/{fileKey}", wapp.DeleteFile).Methods(http.MethodDelete)
	s.HandleFunc("/files/{fileKey}", wapp.UpdateFile).Methods(http.MethodPut).
													  Queries("path", "")
													  // TODO: might want to change something else, not just path.
	s.HandleFunc("/downloadlink/{fileKey}", wapp.WebGetFileDownloadLink).Methods(http.MethodGet)

	// FIXME: passing paths as fileKey's might not be good (need to encode/escape the URL. Might mess parameters up.)

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

// CreateFile API call creates a new file on the cloud.
// Endpoint: /files
// Method: POST.
// Headers: Authorization.
// Query parameters:
// - name=string, the path of the file on the cloud (also the file's key).
// - size=int, the expected size of the file's contents.
// - type=string (optional), the file type (extension).
// - lastModified=date (optional), the date the file was last modified or uploaded.
// Body:
// - File contents as POST body, encoded using post data.
// Response:
// - 200 if file is stored on the cloud successfully on the set path.
func (wapp *webapp) CreateFile(w http.ResponseWriter, req *http.Request) {
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

// ReadFiles API call reads the metadata of all the files that are currently stored on the cloud.
// Endpoint: /files
// Method: GET.
// Headers: Authorization.
// Query parameters: None.
// Response:
// - JSON containing a list of Files.
// - A File contains a key (here, filepath), size, and lastModified time (TODO).
func (wapp *webapp) ReadFiles(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Retrieve files from the cloud.
	files := wapp.cloud.GetFiles()
	utils.GetLogger().Printf("[DEBUG] Got %d files.", len(files))

	// Put into web API file struct format.
	filesWeb := make([]WebFile, 0)
	for _, file := range files {
		webFile := WebFile{
			Key: file.Name,
			Size: int(file.Size),
		}
		filesWeb = append(filesWeb, webFile)
	}
	utils.GetLogger().Printf("[DEBUG] Got %d web files.", len(filesWeb))

	// Serialize as JSON.
	data, err := json.Marshal(filesWeb)
	if err != nil {
		utils.GetLogger().Printf("[ERROR] %v", err)
		w.WriteHeader(http.StatusInternalServerError)
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

// DeleteFile API call deletes a file from the cloud.
// Endpoint: /files/{fileKey}
// - where fileKey is currently the path of the file on the cloud.
// Method: DELETE.
// Headers: Authorization.
// Query parameters: None.
// Response:
// - 200 if file was deleted successfully.
// FIXME: Passing the path as a key might mess things up (forward or backward slashes in URL)
func (wapp *webapp) DeleteFile(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	fileKey := vars["fileKey"]

	filepath := fileKey
	utils.GetLogger().Printf("[DEBUG] Want to delete file with path: %s", filepath)

	locked := wapp.cloud.LockFile(filepath)
	if !locked {
		utils.GetLogger().Printf("[WARN] Could not acquire file lock.")
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	defer wapp.cloud.UnlockFile(filepath)

	err := wapp.cloud.DeleteFile(filepath)
	if err != nil {
		utils.GetLogger().Printf("[ERROR] %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	utils.GetLogger().Printf("[INFO] Deleted file with path: %s", filepath)
	w.WriteHeader(http.StatusOK)
}

// UpdateFile API call updates a file on the cloud.
// Endpoint: /files/{fileKey}
// - where fileKey is currently the path of the file on the cloud.
// Method: POST.
// Headers: Authorization.
// Query parameters:
// - path=string (optional), the new filepath for the file.
// Response:
// - 200 if file was updated successfully.
func (wapp *webapp) UpdateFile(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	fileKey := vars["fileKey"]
	filepath := fileKey

	qs := req.URL.Query()
	utils.GetLogger().Printf("[DEBUG] Querystring parameters: %v", qs)

	newPath, err := GetQueryParam(req.URL, "path")
	if err != nil {
		utils.GetLogger().Printf("[ERROR] %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	utils.GetLogger().Printf("[DEBUG] Want to change %s to %s", filepath, newPath)

	locked := wapp.cloud.LockFile(filepath)
	if !locked {
		utils.GetLogger().Printf("[WARN] Could not acquire file lock: %s", filepath)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	defer wapp.cloud.UnlockFile(filepath)
	locked = wapp.cloud.LockFile(newPath)
	if !locked {
		utils.GetLogger().Printf("[WARN] Could not acquire file lock: %s", newPath)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	defer wapp.cloud.UnlockFile(newPath)

	err = wapp.cloud.MoveFile(filepath, newPath)
	if err != nil {
		utils.GetLogger().Printf("[ERROR] %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

// PingHandler for /ping.
func (wapp *webapp) WebPingHandler(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "pong"}`))
}
