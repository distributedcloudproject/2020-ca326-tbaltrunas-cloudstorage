package datastore

import (
	"cloud/utils"
	"io"
	"errors"
)

type FileSize int

// FileID is a hash as a string of bytes.
type FileID string

// ChunkID is a hash as a string of bytes.
type ChunkID string

type ChunkContents []byte

type FileIOReader io.ReaderAt

// File represents a user's file stored on the cloud.
type File struct {
	ID          FileID  // ID of the file (hash of file contents).

	Path 		string  // Path of the user's file.
	
	Size 		FileSize  // File size.
	
	Chunks 		Chunks  // List of the file's chunk ID's.

	reader		FileIOReader  // Reader used to access the file contents.
}

type Chunks struct {
	NumChunks 	int  // Number of chunks that this file is split into.

	ChunkSize 	int  // The maximum size of each chunk.

	Chunks 		[]Chunk  // List of chunks belonging to the file.
}

// Chunk represents a "chunk" of a file, a sequential part of a file.
// Each chunk has an ID and a sequence number.
type Chunk struct {
	ID 				ChunkID  // Unique ID of the chunk (hash value of the contents).

	SequenceNumber	int // Chunk sequence used to place the chunk in the correct position in the file.

	ContentSize		int // Number of bytes of actual content.
}

// DataStore represents a collection of files.
type DataStore struct {
	Files []*File
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
	allContents := make([]byte, 0)
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
		allContents = append(allContents, buffer...)
		chunks = append(chunks, chunk)
		i++
	}

	// compute file hash
	id := utils.HashFile(allContents)

	// compute extra information
	fileSize := file.Chunks.ComputeFileSize()
	numChunks := len(chunks)
	// Alternatively:
	// numChunks*chunkSize ~= fileSize
	// numChunks ~= ceil(fileSize/chunkSize)

	file.reader = reader
	file.ID = FileID(id)
	file.Path = path
	file.Size = FileSize(fileSize)
	file.Chunks.NumChunks = numChunks
	file.Chunks.ChunkSize = chunkSize
	file.Chunks.Chunks = chunks
	return file, nil
}

// GetChunkByID returns a chunk belonging to the file by its ID.
// Returns nil if the chunk can not be found.
func (file *File) GetChunkByID(chunkID ChunkID) *Chunk {
	for _, chunk := range file.Chunks.Chunks {
		if chunk.ID == chunkID {
			return &chunk
		}
	}
	return nil
}

// Contains returns whether the datastore contains the specified file.
func (ds *DataStore) Contains(file *File) bool {
	for _, f := range ds.Files {
		// TODO: file ID
		if file.Path == f.Path {
			return true
		}
	}
	return false
}

// Add appends a file to the datastore.
func (ds *DataStore) Add(file *File) {
	ds.Files = append(ds.Files, file)
}

// GetChunkByID searches for the chunk with the given ID and the file the chunk belongs to.
// Returns nil if the chunk can not be found.
func (ds *DataStore) GetChunkByID(chunkID ChunkID) (*Chunk, *File) {
	for _, file := range ds.Files {
		chunk := file.GetChunkByID(chunkID)
		if chunk != nil {
			return chunk, file
		}
	}
	return nil, nil
}
