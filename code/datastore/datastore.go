package datastore

import (
	"cloud/network"
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

// // GetChunk returns the contents of the nth chunk in the file.
// func GetChunk(file string, n int) []byte {

// }

// func SplitFile(file string, chunkSize int) n {

// }
