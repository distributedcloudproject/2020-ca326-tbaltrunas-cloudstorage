package network

import (
	"cloud/utils"
	"net/http"
	"encoding/json"

	"golang.org/x/crypto/bcrypt"
)

const (
	accessTokenKey = "access_token"
)

type Credentials struct {
	Username string `json:"username`
	Password string `json:"password"`
}

func (c *cloud) AuthLoginHandler(w http.ResponseWriter, req *http.Request) {
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
	token, expires, err := GenerateToken(creds.Username)
	if err != nil {
		utils.GetLogger().Printf("[ERROR] %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	http.SetCookie(w, &http.Cookie{
		Name: accessTokenKey,
		Value: token,
		Expires: expires,
	})
}

// Adapted from: https://www.sohamkamani.com/blog/golang/2019-01-01-jwt-authentication/
// Should send cookies.
func (c *cloud) AuthRefreshHandler(w http.ResponseWriter, req *http.Request) {
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

	http.SetCookie(w, &http.Cookie{
		Name:    accessTokenKey,
		Value:   newToken,
		Expires: expires,
	})
}
