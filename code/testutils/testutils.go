package testutils

// testutils contains internal utility functions for tests.

import (
	"crypto/rand"
	"crypto/rsa"
)

func GenerateKey() (*rsa.PrivateKey, error) {
	pri, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	return pri, nil
}
