package datastore

import (
	"testing"
	"os"
	"io/ioutil"
	"bytes"
)

func TestFileChunks(t *testing.T) {
	// Create a temporary file with a random string at its end in the default TMP location.
	tmpfile, err := ioutil.TempFile("", "cloud_test_file_*")
	if err != nil { t.Error(err) }
	defer os.Remove(tmpfile.Name())
	defer tmpfile.Close()
	path := tmpfile.Name()
	t.Logf("Temporary filepath: %s", path)

	fileContents := "hellothere"  // 10 bytes
	_, err = tmpfile.Write([]byte(fileContents))
	if err != nil { t.Error(err) }

	fileContentsRead, err := ioutil.ReadFile(path)
	if err != nil { t.Error(err) }
	t.Logf("File contents: %s", string(fileContentsRead))

	chunkSize := 5
	t.Logf("Operating with chunk size: %d", chunkSize)

	r, err := os.Open(path)
	if err != nil { t.Error(err) }
	defer r.Close()
	file, err := NewFile(r, path, chunkSize)
	if err != nil { t.Error(err) }
	t.Logf("File: %v", file)

	// split the file into two chunks
	n := 0
	t.Logf("Chunk: %d", n)
	contents0, bytesRead, err := file.GetChunk(n)
	if err != nil { t.Error(err) }
	t.Logf("Bytes read: %d", bytesRead)
	t.Logf("Contents read: %v (string: %v)", contents0, string(contents0))
	chunkID := ComputeChunkID(contents0)
	if err != nil { t.Error(err) }
	t.Logf("ChunkID: %s", chunkID)

	n = 1
	t.Logf("Chunk: %d", n)
	contents1, bytesRead, err := file.GetChunk(n)
	if err != nil { t.Error(err) }
	t.Logf("Bytes read: %d", bytesRead)
	t.Logf("Contents read: %v (string: %v)", contents1, string(contents1))
	chunkID = ComputeChunkID(contents1)
	if err != nil { t.Error(err) }
	t.Logf("ChunkID: %s", chunkID)

	// collect back the chunks and check against original content
	retrievedContents := string(contents0) + string(contents1)
	t.Logf("Actual contents: %s", fileContents)
	t.Logf("Read contents: %s", retrievedContents)
	if retrievedContents != fileContents {
		t.Errorf("Actual and read contents differ.")
	}
}

func TestChunkSaving(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "cloud_test_file_*")
	if err != nil { t.Error(err) }
	defer os.Remove(tmpfile.Name())
	defer tmpfile.Close()
	path := tmpfile.Name()
	t.Logf("Temporary filepath: %s", path)
	fileContents := "hellothere"  // 10 bytes
	fileContentsBytes := []byte(fileContents)
	t.Logf("Writing contents to temporary file: %s", fileContents)
	_, err = tmpfile.Write(fileContentsBytes)
	if err != nil { t.Error(err) }

	tmpfileSave, err := ioutil.TempFile("", "cloud_test_file_*")
	if err != nil { t.Error(err) }
	defer os.Remove(tmpfileSave.Name())
	defer tmpfileSave.Close()
	pathSave := tmpfileSave.Name()
	t.Logf("Temporary save filepath: %s", pathSave)

	// get a chunk from the file
	chunkSize := 5
	file, err := NewFile(tmpfile, path, chunkSize)
	if err != nil { t.Error(err) }
	t.Logf("File: %v", file)
	chunkNum := 0
	chunk, _, err := file.GetChunk(chunkNum)
	// TODO: do something with the number of bytes read, i.e. store it as a field.
	if err != nil { t.Error(err) }
	t.Logf("Chunk: %v. (string: %s).", chunk, string(chunk))

	// save the chunk
	t.Log("Saving chunk.")
	_, err = file.SaveChunk(tmpfileSave, chunk)
	if err != nil { t.Error(err) }

	// load the chunk
	t.Log("Loading chunk.")
	_, err = tmpfileSave.Seek(0, 0) // reset offset to 0 (since we did a write which moved the pointer)
	if err != nil { t.Error(err) }
	readChunk, _, err := file.LoadChunk(tmpfileSave)
	if err != nil { t.Error(err) }

	// compare read and original chunks
	t.Logf("Original chunk: %v (string: %s).", chunk, string(chunk))
	t.Logf("Read chunk: %v (string: %s).", readChunk, string(readChunk))
	if bytes.Compare(chunk, readChunk) != 0 {
		t.Errorf("Actual and read chunks differ.")
	}
}
