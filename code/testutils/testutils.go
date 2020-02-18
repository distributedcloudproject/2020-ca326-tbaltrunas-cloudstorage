package testutils

// testutils contains internal utility functions for tests.

import (
	"crypto/rand"
	"crypto/rsa"
)

// GenerateKey returns a random RSA private key.
func GenerateKey() (*rsa.PrivateKey, error) {
	pri, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	return pri, nil
}

// RemoveDirs removes all the directories in the list.
func RemoveDirs(dirs []string) {
	for _, dir := range dirs {
		os.RemoveAll(dir)
	}
}
