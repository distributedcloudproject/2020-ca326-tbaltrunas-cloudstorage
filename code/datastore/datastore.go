package datastore

import (
	"cloud/network"
	"io"
	"errors"
)

type FileSize int

// ChunkID is a hash as a string of bytes.
type ChunkID string

type ChunkContents []byte

type ChunkNodeType network.Node

type FileIOReader io.ReaderAt

// File represents a user's file stored on the cloud.
type File struct {
	Path 		string  // Path of the user's file.
	
	Size 		FileSize  // File size.
	
	Chunks 		Chunks  // List of the file's chunk ID's.

	reader		FileIOReader  // Reader used to access the file contents.
}

type Chunks struct {
	NumChunks int  // Number of chunks that this file is split into.

	ChunkSize 	int  // The maximum size of each chunk.

	Chunks []Chunk  // List of chunks belonging to the file.
}

// Chunk represents a "chunk" of a file, a sequential part of a file.
// Each chunk has an ID and a sequence number.
type Chunk struct {
	ID 				ChunkID  // Unique ID of the chunk (hash value of the contents).

	SequenceNumber 	int // Chunk sequence used to place the chunk in the correct position in the file.

	ContentSize       int // Number of bytes of actual content.
}

// ChunkLocations is a data structure that maps from a chunk ID to the Nodes containing that chunk.
// The data structure keeps track of which nodes contain which chunks.
type ChunkLocations map[ChunkID][]ChunkNodeType

// DataStore is a data structure that keeps track of user files stored on the cloud.
type DataStore struct {
	Files [] File
}

// NewFile creates a new File and computes its chunks using the provided chunk size.
// reader is an IO reader that provides access to the underlying file contents.
// path is the expected filepath of the file, used for directory tree purposes.
// chunkSize is the number of bytes that each chunk should be at maximum.
func NewFile(reader FileIOReader, path string, chunkSize int) (*File, error) {
	// validate arguments
	if chunkSize <= 0 {
		return nil, errors.New("Chunk size must be a positive integer.")
	}

	// generate each chunk
	file := new(File)
	chunks := make([]Chunk, 0)
	i := 0
	var offset int64
	buffer := make([]byte, chunkSize)
	stop := false
	for !stop {
		offset = int64(chunkSize * i)
		numRead, err := reader.ReadAt(buffer, offset)
		if err == io.EOF && numRead == 0{
			// EOF and read nothing
			break
		} else if err == io.EOF {
			// EOF but still have something left
			stop = true
		} else if err != nil {
			// other error
			return nil, err
		}
		chunkID := ComputeChunkID(buffer)
		chunk := Chunk{
			ID: chunkID,
			SequenceNumber: i,
			ContentSize: numRead,
		}
		chunks = append(chunks, chunk)
		i++
	}

	// compute extra information
	fileSize := file.Chunks.ComputeFileSize()
	numChunks := len(chunks)
	// Alternatively:
	// numChunks*chunkSize ~= fileSize
	// numChunks ~= ceil(fileSize/chunkSize)

	file.reader = reader
	file.Path = path
	file.Size = FileSize(fileSize)
	file.Chunks.NumChunks = numChunks
	file.Chunks.ChunkSize = chunkSize
	file.Chunks.Chunks = chunks
	return file, nil
}

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
	chunkHash := HashBytes(buffer)
	return ChunkID(chunkHash)
}

// ComputeFileSize calculates the combined size of all chunks (the expected "file size").
func (chunks *Chunks) ComputeFileSize() FileSize {
	fileSize := 0
	for _, chunk := range chunks.Chunks {
		fileSize += chunk.ContentSize
	}
	return FileSize(fileSize)
}

// SaveBytes writes a bytes buffer through a writer.
// It returns the number of bytes actually written.
func (file *File) SaveChunk(w io.Writer, buffer []byte) (int, error) {
	n, err := w.Write(buffer)
	return n, err
}

// LoadBytes reads n bytes from a reader.
// It returns a buffer of the bytes read and the number of actual bytes read.
func (file *File) LoadChunk(r io.Reader) ([]byte, int, error) {
	buffer := make([]byte, file.Chunks.ChunkSize)
	numRead, err := r.Read(buffer)
	if err != nil { return nil, numRead, err }
	return buffer, numRead, nil
}
