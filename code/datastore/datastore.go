package datastore


import (
	"cloud/network"
	"hash"
)


type FileSizeType int
type FileChunkIDType hash.Hash
type FileContentsType [] byte


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
type FileChunkLocations map[FileChunkIDType]network.Node


// DataStore is a data structure for actually storing the user's files on a node.
type DataStore struct {
	Chunks [] FileChunk
}
