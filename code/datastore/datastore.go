package datastore

import (
	"cloud/network"
	"os"
	"io"
	"math"
	"errors"
)

type FileSizeType int

// FileChunkIDType is a hash as a string of bytes.
type FileChunkIDType string

type FileContentsType [] byte

type ChunkNodeType network.Node

// File represents a user's file stored on the cloud.
// Contains the path, file size, and a list of the file's chunk ID's.
type File struct {
	Path 		string
	
	Size 		FileSizeType
	
	Chunks 		FileChunks
}

type FileChunks struct {
	NumberOfChunks int  // Number of chunks that this file is split into.

	ChunkSize 	int  // The maximum size of each chunk.

	Chunks []FileChunk  // List of chunks belonging to the file.
}

// FileChunk represents a "chunk" of a file, a sequential part of a file.
// Each chunk has an ID and a sequence number.
type FileChunk struct {
	ID 				FileChunkIDType  // Unique ID of the chunk (hash value of the contents).

	SequenceNumber 	int // Chunk sequence used to place the chunk in the correct position in the file.
}

// FileChunkLocations is a data structure that maps from a chunk ID to the Nodes containing that chunk.
// The data structure keeps track of which nodes contain which chunks.
type FileChunkLocations map[FileChunkIDType][]ChunkNodeType

// DataStore is a data structure that keeps track of user files stored on the cloud.
type DataStore struct {
	Files [] File
}

// NewFile creates a new File structure from the given filepath and number of chunks.
// If the filepath is invalid, an error is returned.
// This function also splits the file into chunks based on the required number of chunks.
func NewFile(path string, NumberOfChunks int) (*File, error) {
	file := new(File)
	file.Path = path

	size, err := getFileSize(path)
	if err != nil {
		return nil, err
	}
	file.Size = FileSizeType(size)

	file.Chunks.NumberOfChunks = NumberOfChunks
	chunkSize, err := computeChunkSize(int(file.Size), file.Chunks.NumberOfChunks)
	if err != nil {
		return nil, err
	}
	file.Chunks.ChunkSize = chunkSize
	err = file.generateChunks()
	if err != nil {
		return nil, err
	}

	return file, nil
}

// getFileSize computes the size of the file in bytes at the given path.
// Returns 0 and error if file size can not be computed.
func getFileSize(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	fileInfo, err := f.Stat()
	if err != nil {
		return 0, err
	}

	size := int(fileInfo.Size())
	return size, err
}

// computeChunkSize finds the size of each chunk's buffer.
// Chunk size is derived from the given file size and the number of chunks required.
// Returns error if parameters are invalid.
func computeChunkSize(fileSize int, NumberOfChunks int) (int, error) {
	_, ok := interface{}(fileSize).(int)
	if !(0 <= fileSize && ok) {
		return 0, errors.New("File size must be a non-negative integer.")
	}
	_, ok = interface{}(NumberOfChunks).(int)
	if !(0 < NumberOfChunks && ok) {
		return 0, errors.New("Chunk number must be a positive integer.")
	}
	chunkSize := int(math.Ceil(float64(fileSize) / float64(NumberOfChunks)))
	return chunkSize, nil
}

// generateChunks computes and stores the ID (hash) of each chunk in the file.
func (file *File) generateChunks() error {
	file.Chunks.Chunks = make([]FileChunk, file.Chunks.NumberOfChunks)
	for n := 0; n < file.Chunks.NumberOfChunks; n++ {
		chunkID, err := file.GetChunkID(n)
		if err != nil { return err }
		file.Chunks.Chunks[n] = FileChunk{
			ID: chunkID,
			SequenceNumber: n,
		}
	}
	return nil
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

// GetChunkID returns the ID of the nth chunk in the file, or error.
// In implementation this is a hash of the contents of the nth chunk.
// If chunk can not be read, an error is returned.
func (file *File) GetChunkID(n int) (FileChunkIDType, error) {
	buffer, _, err := file.GetChunk(n)
	if err != nil {
		return "", err
	}
	chunkHash := HashBytes(buffer)
	chunkID := FileChunkIDType(chunkHash)
	return chunkID, nil
}
