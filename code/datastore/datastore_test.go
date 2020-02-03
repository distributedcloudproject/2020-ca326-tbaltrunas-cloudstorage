package datastore

import (
	"testing"
	"os"
	"io"
	"math"
	"hash"
	"hash/fnv"
)

func TestGetFileChunk(t *testing.T) {
	path := "/tmp/cloud_test_file"

	f, err := os.Open(path)
	if err != nil {
		t.Error(err)
	}
	defer f.Close()

	// fileInfo, err := f.Stat()
	// if err != nil {
	// 	t.Error(err)
	// }

	// size := fileInfo.Size()
	// t.Logf("File size: %d", size)

	file, err := NewFile(path)
	if err != nil {
		t.Error(err)
	}
	t.Logf("File: %v", file)
	size := file.Size

	chunkNumber := 4
	t.Logf("Operating with chunk number: %d", chunkNumber)

	file.Split(chunkNumber)
	t.Logf("File: %v", file)

	contents, bytesRead, err := file.GetChunk(0)
	t.Logf("Bytes read: %d", bytesRead)
	t.Logf("Contents read: %v (string: %v)", contents, string(contents))
	chunkID, err := file.GetChunkID(0)
	t.Logf("ChunkID: %s", chunkID)

	chunkSize := int(math.Ceil(float64(size)/float64(chunkNumber)))
	t.Logf("Operating with chunk size: %d", chunkSize)

	for	n := 0; n < chunkNumber; n++ {
		t.Logf("Operating on chunk number: %d", n)
		// read chunk
		contents := make([]byte, chunkSize)
		offset := int64(n*chunkSize)
		bytesRead, err := f.ReadAt(contents, offset)
		if err == io.EOF {
			n = chunkNumber + 1
		} else if err != nil {
			t.Error(err)
		}
		t.Logf("Bytes read: %d", bytesRead)
		t.Logf("Contents read: %v (string: %v)", contents, string(contents))

		// hash chunk
		h := hash.Hash(fnv.New32())
		h.Write(contents)
		chunkHash := h.Sum(make([]byte, 0))
		chunkID := FileChunkIDType(chunkHash)
		t.Logf("Hash of contents: %v (string: %v)", chunkHash, chunkID)
	}
}
