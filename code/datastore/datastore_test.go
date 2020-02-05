package datastore

import (
	"testing"
	"os"
	"io/ioutil"
	"bytes"
)

// CreateTestFile utility function creates a file with contents.
// An error is returned if the operation can no be performed.
func CreateTestFile(path string, contents string) error {
	err := ioutil.WriteFile(path, []byte(contents), os.ModePerm)
	return err
}

func TestFileSplit(t *testing.T) {
	// create a file with test contents
	path := "/tmp/cloud_test_file"
	t.Logf("File path: %s", path)

	fileContents := "hellothere"  // 10 bytes
	err := CreateTestFile(path, fileContents)
	if err != nil {
		t.Error(err)
	}
	fileContentsRead, err := ioutil.ReadFile(path)
	if err != nil {
		t.Error(err)
	}
	t.Logf("File contents: %s", string(fileContentsRead))

	chunkNumber := 2
	t.Logf("Operating with chunk number: %d", chunkNumber)

	file, err := NewFile(path, chunkNumber)
	if err != nil {
		t.Error(err)
	}
	t.Logf("File: %v", file)

	// split the file into two chunks
	n := 0
	t.Logf("Chunk: %d", n)
	contents0, bytesRead, err := file.GetChunk(n)
	if err != nil {
		t.Error(err)
	}
	t.Logf("Bytes read: %d", bytesRead)
	t.Logf("Contents read: %v (string: %v)", contents0, string(contents0))
	chunkID, err := file.GetChunkID(n)
	if err != nil {
		t.Error(err)
	}
	t.Logf("ChunkID: %s", chunkID)

	n = 1
	t.Logf("Chunk: %d", n)
	contents1, bytesRead, err := file.GetChunk(n)
	if err != nil {
		t.Error(err)
	}
	t.Logf("Bytes read: %d", bytesRead)
	t.Logf("Contents read: %v (string: %v)", contents1, string(contents1))
	chunkID, err = file.GetChunkID(n)
	if err != nil {
		t.Error(err)
	}
	t.Logf("ChunkID: %s", chunkID)

	// collect back the chunks and check against original content
	retrievedContents := string(contents0) + string(contents1)
	t.Logf("Actual contents: %s", fileContents)
	t.Logf("Read contents: %s", retrievedContents)
	if retrievedContents != fileContents {
		t.Errorf("Actual and read contents differ.")
	}
}

func TestFileSave(t *testing.T) {
	// get a file
	path := "/tmp/cloud_test_file"
	fileContents := "hellothere"
	err := CreateTestFile(path, fileContents)
	if err != nil { t.Error(err) }

	// get a chunk from the file
	chunkNumber := 2
	file, err := NewFile(path, chunkNumber)
	if err != nil { t.Error(err) }
	t.Logf("File: %v", file)

	// save the chunk
	n := 0
	chunkPath := "/tmp/cloud_test_chunk_save"

	err = file.SaveChunk(n, chunkPath)
	if err != nil { t.Error(err) }

	// retrieve and compare the chunk
	chunk, _, err := file.GetChunk(n)
	if err != nil { t.Error(err) }
	readChunk, err := file.LoadChunk(chunkPath)
	if err != nil { t.Error(err) }

	t.Logf("Actual chunk: %v", chunk)
	t.Logf("Read chunk: %v", readChunk)
	if bytes.Compare(chunk, readChunk) != 0 {
		t.Errorf("Actual and read chunks differ.")
	}
}
