package datastore


import (
	"cloud/network"
	"hash"
)


type FileSizeType int
type FileChunkIDType hash.Hash
type FileContentsType [] byte


// a user file stored on the cloud
type File struct {
	Path 		string
	
	Size 		FileSizeType
	
	ChunkIDs [] FileChunkIDType  // list of chunks belonging to the file
}


// a part of a file
type FileChunk struct {
	ID 				FileChunkIDType  // unique id of the chunk (hash value of the contents)

	SequenceNumber 	int // chunk sequence inside a file it belongs to

	Contents 		FileContentsType // actual contents of the chunk
}


// data structure that keeps track of which nodes contain which chunks
type FileChunkLocations map[FileChunkIDType]network.Node


// data structure for actually storing user files on a node
type DataStore struct {
	Chunks [] FileChunk
}
