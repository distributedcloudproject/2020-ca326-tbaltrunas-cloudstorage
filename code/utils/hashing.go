package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

// HashFile computes the hash of a file's bytes using a suitable hash function.
// The function returns the hash as a hex-encoded string.
func HashFile(buffer []byte) string {
	// Adapted from function network.PublicKeyToID.
	hash := sha256.Sum256(buffer)
	result := hex.EncodeToString(hash[:])
	return result
}
