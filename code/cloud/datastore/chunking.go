package datastore

import (
	"cloud/utils"
	"io"
	"errors"
	"fmt"
)

// GetChunk reads the nth chunk in the file.
// Returns the contents as bytes, the amount of actual bytes read, and error if any.
func (file *File) GetChunk(n int) ([]byte, int, error) {
	offset := int64(n * file.Chunks.ChunkSize)
	buffer := make([]byte, file.Chunks.ChunkSize)
	numRead, err := file.reader.ReadAt(buffer, offset)
	if err != io.EOF && err != nil { return nil, numRead, err }
	return buffer, numRead, nil
	// TODO: might want to do something with numRead, i.e. update chunk with new ContentSize and ID.
}

// ComputeChunkID calculates the ID (hash) of a buffer of bytes (a chunk).
func ComputeChunkID(buffer []byte) ChunkID {
	chunkHash := utils.HashFile(buffer)
	return ChunkID(chunkHash)
}

// ComputeFileSize calculates the combined size of all chunks (the expected "file size").
func (chunks *Chunks) ComputeFileSize() uint64 {
	var fileSize uint64 = 0
	for _, chunk := range chunks.Chunks {
		fileSize += chunk.ContentSize
	}
	return fileSize
}

// SaveChunk writes a bytes buffer through a writer, until the buffer is fully written.
func SaveChunk(w io.Writer, buffer []byte) error {
	written := 0
	for written < len(buffer) {
		n, err := w.Write(buffer[written:])
		written += n
		if err != nil {
			return err
		}
	}
	return nil
}

// LoadChunk reads a chunk from a reader.
func (file *File) LoadChunk(r io.Reader) ([]byte, error) {
	buffer := make([]byte, file.Chunks.ChunkSize)
	numRead, err := r.Read(buffer)
	if numRead != file.Chunks.ChunkSize {
		return nil, errors.New(fmt.Sprintf("Chunk requires %d bytes. Read %d bytes", file.Chunks.ChunkSize, numRead))
	} else if err != io.EOF && err != nil {
		return nil, err 
	}
	return buffer, nil
}
