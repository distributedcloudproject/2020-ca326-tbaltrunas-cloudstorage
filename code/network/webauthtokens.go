package network

import (
	"os"
	"time"
	"errors"

	"github.com/dgrijalva/jwt-go"
)

const (
	accessTokenExpirationTime = 60 * time.Minute
	refreshMinTimeLeft = 30 * time.Second
)

var (
	jwtKey = []byte(os.Getenv("JWT_KEY"))
)

// authClaims is sent as part of a JWT for standard user claims.
type authClaims struct {
	Username string  `json:"username"` // Unique username for this token.
	jwt.StandardClaims  // includes expiry time
}

func jwtKeyFunc(token *jwt.Token) (interface{}, error) {
	return jwtKey, nil
}

func GenerateToken(username string) (string, time.Time, error) {
	expirationTime := time.Now().Add(accessTokenExpirationTime)

	claims := authClaims{
		Username: username, // TODO: roles based on username
		StandardClaims: jwt.StandardClaims {
			ExpiresAt: expirationTime.Unix(),
		},
	}

	// TODO: decide on the signing method
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", expirationTime, err
	}
	return tokenString, expirationTime, nil
}

func RefreshToken(token string) (string, time.Time, error) {
	// Note that token wwill be already verified by the authentication middleware.

	// Get claims of existing token including expiration time.
	var expires time.Time
	claims := &authClaims{}
	_, err := jwt.ParseWithClaims(token, claims, jwtKeyFunc)
	if err != nil {
		return "", expires, err
	}

	// New token is only issued if the expiry time of the current token is within a certain limit.
	if time.Unix(claims.ExpiresAt, 0).Sub(time.Now()) > refreshMinTimeLeft {
		return "", expires, errors.New("Provided token not expired")
	}

	// Generate a completely new token.
	return GenerateToken(claims.Username)
}

func ValidateToken(token string) error {
	claims := &authClaims{}
	parsedToken, err := jwt.ParseWithClaims(token, claims, jwtKeyFunc)
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return errors.New("Signature invalid")
		} else {
			return err
		}
	}
	if !parsedToken.Valid {
		return errors.New("Signature invalid")
	}
	return nil
}
