package network

import (
	// "cloud/datastore"
	"cloud/utils"
	"net/http"
	"strconv"
	"fmt"
	"encoding/json"
	"time"

	"github.com/gorilla/mux"
	gorillaHandlers "github.com/gorilla/handlers"
	"github.com/dgrijalva/jwt-go"
)

var jwtSecret = []byte("bigsecret")

type File struct {
	Key		string   `json:"key"`
}

// AuthClaims is sent as part of a JWT for standard user claims.
type AuthClaims struct {
	Username string  `json:"username"` // Unique username for this token.
	jwt.StandardClaims  // includes expiry time
}

// ListenAndServeHTTP starts a HTTP server and routes requests to the cloud.
func (c *cloud) ListenAndServeHTTP(port int) error {
	address := ":" + strconv.Itoa(port)

	r := mux.NewRouter()
	r.HandleFunc("/auth", c.WebAuthenticationHandler)

	r.HandleFunc("/netinfo", c.NetworkInfoHandler)
	r.HandleFunc("/files", c.GetFiles).Methods(http.MethodGet)
	r.HandleFunc("/files/{fileID}", c.GetFile).
		Methods(http.MethodGet).
		Queries("filter", "contents")
	r.HandleFunc("/files", c.CreateFile).Methods(http.MethodPost)


	http.Handle("/", r)

	// Set up CORS middleware for local development.
	originsOk := gorillaHandlers.AllowedOrigins([]string{"http://localhost"})
	methodsOk := gorillaHandlers.AllowedMethods([]string{http.MethodOptions, http.MethodGet, http.MethodPost})

	utils.GetLogger().Printf("[INFO] HTTP backend listening on address: %s.", address)
	return http.ListenAndServe(address, gorillaHandlers.CORS(originsOk, methodsOk)(r))
}

func (c *cloud) WebAuthenticationHandler(w http.ResponseWriter, req *http.Request) {
	utils.GetLogger().Println("[INFO] WebAuthenticationHandler called.")

	// TODO: verify sent request, i.e. username and password

	// expires in 5 minutes
	expirationTime := time.Now().Add(1 * time.Minute)
	authClaims := AuthClaims{
		Username: "admin", // TODO: roles based on username
		StandardClaims: jwt.StandardClaims {
			ExpiresAt: expirationTime.Unix(), // represent expiration time as Unix milliseconds
		},
	}
	// TODO: decide on the signing method, i.e. use ssh key
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, authClaims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		utils.GetLogger().Printf("[ERROR] %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name: "access_token",
		Value: tokenString,
		Expires: expirationTime,
	})
}

func (c *cloud) NetworkInfoHandler(w http.ResponseWriter, req *http.Request) {
	utils.GetLogger().Println("[INFO] NetworkInfoHandler called.")
	w.WriteHeader(http.StatusOK)
	networkName := c.Network().Name
	w.Write([]byte(fmt.Sprintf(`{"name": "%s"}`, networkName)))
}

// Required Body:
// Required Query Params:
// Optional Query Params:
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
