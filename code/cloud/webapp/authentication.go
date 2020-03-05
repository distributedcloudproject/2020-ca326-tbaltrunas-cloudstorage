package webapp

import (
	"cloud/utils"
	"net/http"
	"encoding/json"
	"strings"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

const (
	accessTokenKey = "access_token"
)

type Credentials struct {
	Username string `json:"username`
	Password string `json:"password"`
}

func (wapp *webapp) AuthLogin(w http.ResponseWriter, req *http.Request) {
	var creds Credentials
	err := json.NewDecoder(req.Body).Decode(&creds)
	if err != nil {
		utils.GetLogger().Printf("[ERROR] %v", err)
		// TODO: include error messages in bad response bodies.
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	utils.GetLogger().Printf("[DEBUG] Credentials: %v", creds.Username)
	
	// Check username and get password of the account.
	storedAccount, err := DBGetAccountByUsername(creds.Username)
	if err != nil {
		utils.GetLogger().Printf("[ERROR] %v", err)
		if err.Error() == "DB connection error" {
			// Could not connect to DB.
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			// Record not found.
			w.WriteHeader(http.StatusUnauthorized)
		}
		return
	}
	utils.GetLogger().Printf("[DEBUG] Stored Account: %v", storedAccount.Username)

	// Check password
	err = bcrypt.CompareHashAndPassword([]byte(storedAccount.Password), []byte(creds.Password))
	if err != nil {
		// Comparison failed
		utils.GetLogger().Printf("[ERROR] %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Success. Generate an access token.
	token, expires, err := GenerateAccessToken(creds.Username)
	if err != nil {
		utils.GetLogger().Printf("[ERROR] %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	http.SetCookie(w, &http.Cookie{
		Name: accessTokenKey,
		Value: token,
		Expires: expires,
		Path: "/",
	})
}

// Adapted from: https://www.sohamkamani.com/blog/golang/2019-01-01-jwt-authentication/
// Should send cookies.
func (wapp *webapp) AuthRefresh(w http.ResponseWriter, req *http.Request) {
	cookie, err := req.Cookie(accessTokenKey)
	if err != nil {
		utils.GetLogger().Printf("[ERROR] %v", err)
		if err == http.ErrNoCookie {
			// Caller did not send the cookie with the token.
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
		return
	}

	newToken, expires, err := RefreshToken(cookie.Value)
	if err != nil {
		utils.GetLogger().Printf("[ERROR] %v", err)
		if err.Error() == "Provided token not expired" {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	// FIXME: expiration time doesn't actually do anything (doesn't expire)

	// TODO: Delete old cookie.
	http.SetCookie(w, &http.Cookie{
		Name:    accessTokenKey,
		Value:   newToken,
		Expires: expires,
		Path: "/",
	})
}

// Adapted from: https://blog.usejournal.com/authentication-in-golang-c0677bcce1a8
func AuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get token from the header
		authType, credentials, err := parseAuthorizationHeader(r.Header.Get("Authorization"))
		if err != nil {
			utils.GetLogger().Printf("[ERROR] %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if authType != "Bearer" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		token := credentials
		if token == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		err = ValidateAccessToken(token)
		if err != nil {
			utils.GetLogger().Printf("[ERROR] %v", err)
			if err.Error() == "Signature invalid" {
				w.WriteHeader(http.StatusUnauthorized)
			} else {
				w.WriteHeader(http.StatusBadRequest)
			}
			return
		}
		next.ServeHTTP(w, r)
	})
}

func parseAuthorizationHeader(headerValue string) (string, string, error) {
	splitRes := strings.Split(headerValue, " ")
	if len(splitRes) != 2 {
		return "", "", errors.New("Malformed header")
	}
	authType, credentials := splitRes[0], splitRes[1]
	credentials = strings.TrimSpace(credentials)
	return authType, credentials, nil
}
