package utils

import (
	"hash"
	"hash/fnv"
)


// TODO: create a struct/interface for choosing the default hash function to use.
var hashFunction = hash.Hash(fnv.New32())

// HashBytes hashes a byte slice using a default hashing implementation.
// The function returns the hash result as a byte slice.
func HashBytes(buffer []byte) []byte {
	hashFunction.Write(buffer)
	result := hashFunction.Sum(make([]byte, 0))
	return result
}
