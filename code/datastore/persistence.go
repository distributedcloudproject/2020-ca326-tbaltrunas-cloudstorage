package datastore

import (
	"os"
	"io/ioutil"
	"bytes"
	"encoding/gob"
)

// Save persistently stores the struct s into a file at path as bytes.
// The implementation follows the following tutorial:
// https://medium.com/@matryer/golang-advent-calendar-day-eleven-persisting-go-objects-to-disk-7caf1ee3d11d
// Unlike in the tutorial, we use ioutil for IO and gob for encoding.
// Use the function as follows:
// Save('/path', structure)
func Save(path string, s interface{}) error {
	serialized, err := Encode(s)
	if err != nil {
		return err
	}
	// os.ModePerm is Unix permissions for everything (0o777)
	ioutil.WriteFile(path, serialized, os.ModePerm)
	return nil
}

// Load decodes bytes at the filepath into the struct s.
// Note that s must be a reference.
// For example, use as follows:
// Load('/path', &structure)
func Load(path string, s interface{}) error {
	// read file into buffer
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	// decode structure from bytes
	err = Decode(contents, s)
	return nil
}

// Encode serializes the struct s into an array of bytes.
func Encode(s interface{}) ([]byte, error) {
	var buffer bytes.Buffer
	// FIXME: for performance reasons reuse enc across function calls?
	enc := gob.NewEncoder(&buffer)
	err := enc.Encode(s)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// Decode deserializes the buffer of bytes into the struct variable s.
// s must be a reference
// For example, use as:
// Decode(buf, &structure)
func Decode(buffer []byte, s interface{}) error {
	b := bytes.NewBuffer(buffer)
	// FIXME: for performance reasons reuse dec across function calls?
	dec := gob.NewDecoder(b)
	err := dec.Decode(s)
	if err != nil {
		return err
	}
	return nil
}
