package datastore


import (
	"cloud/network"
	"hash"
)


type FileSizeType int
type FileChunkIdType hash.Hash
type FileContentsType [] byte


// a user file stored on the cloud
type File struct {
	Path 		string
	
	Size 		FileSizeType
	
	ChunkIds [] FileChunkIdType  // list of chunks belonging to the file
}


// a part of a file
type FileChunk struct {
	id 				FileChunkIdType  // unique id of the chunk

	sequenceNumber 	int // chunk sequence inside a file it belongs to

	contents 		FileContentsType // actual contents of the chunk
}


// data structure that keeps track of which nodes contain which chunks
type FileChunkLocations map[FileChunkIdType]network.Node


// data structure for actually storing user files on a node
type DataStore struct {
	chunks [] FileChunk
}
