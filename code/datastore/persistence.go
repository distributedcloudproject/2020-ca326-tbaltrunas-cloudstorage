package datastore

import (
	"os"
	"io/ioutil"
	"bytes"
	"encoding/gob"
)

// Save persistently stores the struct s into a file at path as bytes.
// The implementation closely follows the following tutorial:
// https://medium.com/@matryer/golang-advent-calendar-day-eleven-persisting-go-objects-to-disk-7caf1ee3d11d
// Except that the encoding format used is gob.
func Save(path string, s interface{}) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	// encode s as bytes
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	err = enc.Encode(s)
	if err != nil {
		return err
	}
	// write buffer into file
	f.Write(buffer.Bytes())
	return nil
}

// Load decodes bytes at the filepath into the struct s.
func Load(path string, s interface{}) error {
	// read file into buffer
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	var buffer bytes.Buffer
	buffer.Write(contents)
	// decode buffer into s
	dec := gob.NewDecoder(&buffer)
	err = dec.Decode(s)
	if err != nil {
		return err
	}
	return nil
}
