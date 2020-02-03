package datastore

import (
	"cloud/network"
	"os"
	"io"
	"math"
	"hash"
	"hash/fnv"
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
	ChunkNumber int  // Number of chunks that this file is split into

	ChunkSize 	int  // The maximum size of each chunk

	ChunkIDs [] FileChunkIDType  // List of chunks belonging to the file.
}

// FileChunk represents a "chunk" of a file, a sequential part of a file.
// Each chunk has an ID, a sequence number, and actual contents of the chunk.
type FileChunk struct {
	ID 				FileChunkIDType  // Unique ID of the chunk (hash value of the contents).

	SequenceNumber 	int // Chunk sequence used to place the chunk in the correct position in the file.

	Contents 		FileContentsType
}

// FileChunkLocations is a data structure that maps from a chunk ID to the Nodes containing that chunk.
// The data structure keeps track of which nodes contain which chunks.
type FileChunkLocations map[FileChunkIDType][]ChunkNodeType

// DataStore is a data structure that keeps track of user files stored on the cloud.
type DataStore struct {
	Files [] File
}

// NewFile creates a new File structure from the given filepath.
// If the filepath is invalid, an error is returned.
func NewFile(path string) (*File, error) {
	file := new(File)
	file.Path = path

	size, err := fileSize(path)
	file.Size = FileSizeType(size)

	return file, err
}

// fileSize computes the size of the file in bytes at the given path.
// Returns 0 and error if file size can not be computed.
func fileSize(path string) (int, error) {
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

// GetChunk returns the contents of the nth chunk in the file.
// func GetChunk(file string, n int) []byte {

// }

// Split divides up the current file into chunks.
// By calculating the required offsets and hashes for the File.
func (file *File) Split(chunkNumber int) {
	file.Chunks.ChunkNumber = chunkNumber

	chunkSize := int(math.Ceil(float64(file.Size)/float64(chunkNumber)))
	file.Chunks.ChunkSize = chunkSize


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
func (file *File) GetChunkID(n int) (FileChunkIDType, error) {
	buffer, _, err := file.GetChunk(n)
	if err != nil {
		return "", err
	}
	h := hash.Hash(fnv.New32())
	h.Write(buffer)
	chunkHash := h.Sum(make([]byte, 0))
	chunkID := FileChunkIDType(chunkHash)
	return chunkID, nil
}
