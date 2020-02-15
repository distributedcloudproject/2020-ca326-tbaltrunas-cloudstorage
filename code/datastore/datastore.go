package datastore

import (
	"cloud/network"
	"os"
	"io"
	"errors"
)

type FileSizeType int

// FileChunkIDType is a hash as a string of bytes.
type FileChunkIDType string

type FileContentsType [] byte

type ChunkNodeType network.Node

type FileIOReader io.ReaderAt

// File represents a user's file stored on the cloud.
type File struct {
	Path 		string  // Path of the user's file.
	
	Size 		FileSizeType  // File size.
	
	Chunks 		FileChunks  // List of the file's chunk ID's.

	reader		FileIOReader  // Reader used to access the file contents.
}

type FileChunks struct {
	NumChunks int  // Number of chunks that this file is split into.

	ChunkSize 	int  // The maximum size of each chunk.

	Chunks []FileChunk  // List of chunks belonging to the file.
}

// FileChunk represents a "chunk" of a file, a sequential part of a file.
// Each chunk has an ID and a sequence number.
type FileChunk struct {
	ID 				FileChunkIDType  // Unique ID of the chunk (hash value of the contents).

	SequenceNumber 	int // Chunk sequence used to place the chunk in the correct position in the file.

	ContentSize       int // Number of bytes of actual content.
}

// FileChunkLocations is a data structure that maps from a chunk ID to the Nodes containing that chunk.
// The data structure keeps track of which nodes contain which chunks.
type FileChunkLocations map[FileChunkIDType][]ChunkNodeType

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
	chunks := make([]FileChunk, 0)
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
		chunk := FileChunk{
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
	file.Size = FileSizeType(fileSize)
	file.Chunks.NumChunks = numChunks
	file.Chunks.ChunkSize = chunkSize
	file.Chunks.Chunks = chunks
	return file, nil
}

// GetChunk reads the nth chunk in the file.
// Returns the contents as bytes, the amount of actual bytes read, and error if any.
func (file *File) GetChunk(n int) ([]byte, int, error) {
	f, err := os.Open(file.Path)
	if err != nil {
		return nil, 0, err
	}
	defer f.Close()

	buffer := make([]byte, file.Chunks.ChunkSize)
	offset := int64(n * file.Chunks.ChunkSize)
	bytesRead, err := f.ReadAt(buffer, offset)
	if err != io.EOF && err != nil {
		return nil, 0, err
	}
	return buffer, bytesRead, nil
}

func ComputeChunkID(buffer []byte) (FileChunkIDType) {
	chunkHash := HashBytes(buffer)
	return FileChunkIDType(chunkHash)
}

func (chunks *FileChunks) ComputeFileSize() FileSizeType {
	fileSize := 0
	for _, chunk := range chunks.Chunks {
		fileSize += chunk.ContentSize
	}
	return FileSizeType(fileSize)
}
