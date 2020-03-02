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

type File struct {
	Key		string   `json:"key"`
}

// ServeWebApp starts a backend web application server with the appropriate routers.
func (c *cloud) ServeWebApp(port int) error {
	address := ":" + strconv.Itoa(port)

	r := mux.NewRouter()

	// Do not need auth
	r.HandleFunc("/ping", c.PingHandler)
	r.HandleFunc("/auth/login", c.AuthLoginHandler).Methods(http.MethodPost)

	// Need auth
	s := r.PathPrefix("/").Subrouter()
	s.Use(AuthenticationMiddleware)
	s.HandleFunc("/auth/refresh", c.AuthRefreshHandler)
	s.HandleFunc("/netinfo", c.NetworkInfoHandler)
	s.HandleFunc("/files", c.GetFiles).Methods(http.MethodGet)
	s.HandleFunc("/files/{fileID}", c.GetFile).
		Methods(http.MethodGet).
		Queries("filter", "contents")
	s.HandleFunc("/files", c.CreateFile).Methods(http.MethodPost)

	http.Handle("/", r)

	// Set up CORS middleware for local development.
	originsOk := gorillaHandlers.AllowedOrigins([]string{"http://localhost"})
	methodsOk := gorillaHandlers.AllowedMethods([]string{http.MethodOptions, http.MethodGet, http.MethodPost})

	utils.GetLogger().Printf("[DEBUG] Cert and key: %s, %s", os.Getenv("SSL_CRT_FILE"), os.Getenv("SSL_KEY_FILE"))
	utils.GetLogger().Printf("[INFO] Web backend listening on address: %s.", address)
	return http.ListenAndServeTLS(address, os.Getenv("SSL_CRT_FILE"), os.Getenv("SSL_KEY_FILE"), 
			gorillaHandlers.LoggingHandler(os.Stdout, 
				gorillaHandlers.CORS(originsOk, methodsOk)(r)))
	// TODO: use utils.GetLogger() writer
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

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.Itoa(fileLength))
	w.Header().Set("Content-Disposition", 
				   fmt.Sprintf("attachment;filename=%s", fileName))
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
