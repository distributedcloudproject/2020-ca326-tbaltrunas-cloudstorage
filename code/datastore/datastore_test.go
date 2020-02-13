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

	NumChunks := 2
	t.Logf("Operating with chunk number: %d", NumChunks)

	file, err := NewFileNumChunks(path, NumChunks)
	if err != nil { t.Error(err) }
	t.Logf("File: %v", file)

	// split the file into two chunks
	n := 0
	t.Logf("Chunk: %d", n)
	contents0, bytesRead, err := file.GetChunk(n)
	if err != nil { t.Error(err) }
	t.Logf("Bytes read: %d", bytesRead)
	t.Logf("Contents read: %v (string: %v)", contents0, string(contents0))
	chunkID, err := file.GetChunkID(n)
	if err != nil { t.Error(err) }
	t.Logf("ChunkID: %s", chunkID)

	n = 1
	t.Logf("Chunk: %d", n)
	contents1, bytesRead, err := file.GetChunk(n)
	if err != nil { t.Error(err) }
	t.Logf("Bytes read: %d", bytesRead)
	t.Logf("Contents read: %v (string: %v)", contents1, string(contents1))
	chunkID, err = file.GetChunkID(n)
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

func TestNewFile(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "cloud_test_file_*")
	if err != nil { t.Error(err) }
	defer os.Remove(tmpfile.Name())
	defer tmpfile.Close()

	path := tmpfile.Name()
	t.Logf("Temporary filepath: %s", path)

	fileContents := "hellothere"  // 10 bytes
	_, err = tmpfile.Write([]byte(fileContents))
	if err != nil { t.Error(err) }

		file1, err := NewFileNumChunks(path, 2)  // 2 chunks of 5 bytes each
	if err != nil { t.Error(err) }
	t.Logf("File from NumChunks (File 1): %v.", file1)

	file2, err := NewFileChunkSize(path, 5)  // 5 bytes giving 2 chunks
	if err != nil { t.Error(err) }
	t.Logf("File from ChunkSize (File 2): %v.", file2)

	t.Logf("File 1 number of chunks: %d.", file1.Chunks.NumChunks)
	t.Logf("File 2 number of chunks: %d.", file2.Chunks.NumChunks)
	if file1.Chunks.NumChunks != file2.Chunks.NumChunks {
		t.Error("NumChunks does not match.")
	}

	t.Logf("File 1 chunk size: %d.", file1.Chunks.ChunkSize)
	t.Logf("File 2 chunk size: %d.", file2.Chunks.ChunkSize)
	if file1.Chunks.ChunkSize != file2.Chunks.ChunkSize {
		t.Error("ChunkSize does not match.") 
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
	numChunks := 2
	file, err := NewFileNumChunks(path, numChunks)
	if err != nil { t.Error(err) }
	t.Logf("File: %v", file)
	chunkNum := 0
	chunk, _, err := file.GetChunk(chunkNum)
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
