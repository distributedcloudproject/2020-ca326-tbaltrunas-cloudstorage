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

func TestFileSave(t *testing.T) {
	// get a file
	tmpfile, err := ioutil.TempFile("", "cloud_test_file_*")
	if err != nil { t.Error(err) }
	defer os.Remove(tmpfile.Name())
	defer tmpfile.Close()

	path := tmpfile.Name()
	t.Logf("Temporary filepath: %s", path)

	fileContents := "hellothere"  // 10 bytes
	_, err = tmpfile.Write([]byte(fileContents))
	if err != nil { t.Error(err) }

	// get a chunk from the file
	chunkNumber := 2
	file, err := NewFileNumChunks(path, chunkNumber)
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
