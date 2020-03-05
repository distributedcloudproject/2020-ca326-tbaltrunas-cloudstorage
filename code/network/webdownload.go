package network

import (
	"cloud/utils"
	"net/http"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/gorilla/mux"
)

// FileDownloadLink API call generates a public (unauthenticated) temporary download link for a file.
// Endpoint: /downloadlink/{fileKey}
// - where fileKey is the filepath of the file to be downloaded.
// Method: GET.
// Headers: Authorization.
// Query parameters: None.
// Body: None.
// Response:
// - The body of the request will contain the file download link (endpoint + token as a query string parameter).
// Approach based on: https://codeburst.io/part-1-jwt-to-authenticate-downloadable-files-at-client-8e0b979c9ac1
func (wapp *webapp) FileDownloadLink(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	fileID := vars["fileKey"]

	token, err := GenerateDownloadToken(fileID)
	if err != nil {
		utils.GetLogger().Printf("[ERROR] %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	fileURL := fmt.Sprintf("/downloadfile/%s?token=%s", fileID, token)
	w.Write([]byte(fileURL))
}

// DownloadFile API call returns a stream of file contents with suitable headers to initiate client side download.
// Endpoint: /downloadfile/{fileKey}
// - where fileKey is the filepath of the file to be downloaded.
// Method: GET.
// Headers: None (public route).
// Query parameters:
// - token, the secret temporary token that authenticates the user and allows them to download the file.
// Body: None.
// Response:
// - The file as an octect (byte) stream with the suitable browser headers.
func (wapp *webapp) DownloadFile(w http.ResponseWriter, req *http.Request) {
	// token should be verified by middleware
	vars := mux.Vars(req)
	fileKey := vars["fileKey"]
	filepath := fileKey

	// Get the file by key for extra information
	file, err := wapp.cloud.GetFile(filepath)
	if err != nil {
		utils.GetLogger().Printf("[ERROR] %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Create a temporary file where to store the downloaded file
	tmpFile, err := ioutil.TempFile("", "cloud_dl_file*")
	defer os.Remove(tmpFile.Name())
	if err != nil {
		utils.GetLogger().Printf("[ERROR] %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	localFile := tmpFile.Name()
	tmpFile.Close()

	utils.GetLogger().Printf("[DEBUG] Initiated a download manager for file: %v", filepath)

	// Download the file on the web app server
	dm := DownloadManager{
		Cloud: wapp.cloud.(*cloud),
	}
	dm.downloadFile(filepath, localFile)
	// FIXME: function should probably be public (start with uppercase)
	// FIXME: download token likely to expire if download takes a long time on the cloud side

	// Open downloaded contents
	fileContents, err := os.Open(localFile)
	if err != nil {
		utils.GetLogger().Printf("[ERROR] %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer fileContents.Close()

	// Set headers on response to suggest download on client-side
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.Itoa(int(file.Size)))
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=%s", filepath))

	// Write out contents to response
	buffer := make([]byte, wapp.cloud.Config().FileChunkSize)
	written := 0
	for written < int(file.Size) {
		numRead, err := fileContents.Read(buffer)
		if err != nil && err != io.EOF {
			utils.GetLogger().Printf("[ERROR] %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		numWritten, err := w.Write(buffer[:numRead])
		if err != nil {
			utils.GetLogger().Printf("[ERROR] %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		written += numWritten
	}
	// FIXME: reuse this pattern of read-write as a function
}

// DownloadTokenMiddleware verifies that a download token is valid before passing on the request to a downloader.
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
