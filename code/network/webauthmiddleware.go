package network

import (
	"cloud/utils"

	"net/http"
	"strings"
	"errors"
)

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
