package datastore

import (
	"testing"
	"os"
	"io/ioutil"
)

func TestFileChunks(t *testing.T) {
	// create a file with test contents
	path := "/tmp/cloud_test_file"
	t.Logf("File path: %s", path)

	fileContents := "hellothere"  // 10 bytes
	err := ioutil.WriteFile(path, []byte(fileContents), os.ModePerm)
	if err != nil {
		t.Error(err)
	}
	fileContentsRead, err := ioutil.ReadFile(path)
	if err != nil {
		t.Error(err)
	}
	t.Logf("File contents: %s", string(fileContentsRead))

	NumberOfChunks := 2
	t.Logf("Operating with chunk number: %d", NumberOfChunks)

	file, err := NewFile(path, NumberOfChunks)
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
