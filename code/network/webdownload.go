package network

import (
	"cloud/utils"
	"net/http"
	"fmt"
	"strconv"

	"github.com/gorilla/mux"
)

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
		return
	}
	if n != fileLength {
		utils.GetLogger().Printf("[WARN] Written file length does not match: %d (want %d)", n, fileLength)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
