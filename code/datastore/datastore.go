package datastore

import (
	"cloud/utils"
	"errors"
	"io"
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
	ID FileID // ID of the file (hash of chunk IDs).

	Name string // The name of the file.

	Size FileSize // File size.

	Chunks Chunks // List of the file's chunk ID's.

	reader FileIOReader // Reader used to access the file contents.
}

type Chunks struct {
	NumChunks int // Number of chunks that this file is split into.

	ChunkSize int // The maximum size of each chunk.

	Chunks []Chunk // List of chunks belonging to the file.
}

// Chunk represents a "chunk" of a file, a sequential part of a file.
// Each chunk has an ID and a sequence number.
type Chunk struct {
	ID ChunkID // Unique ID of the chunk (hash value of the contents).

	SequenceNumber int   // Chunk sequence used to place the chunk in the correct position in the file.
	ContentSize    int   // Number of bytes of actual content.
	ChunkOffset    int64 // The offset of the chunk of the file. (e.g, if offset is 1024, the chunk is contained at
	// position 1024 of the file.
}

// ChunkStore represents a locally stored Chunk.
type ChunkStore struct {
	Chunk  Chunk
	FileID FileID

	// StoredAsFile states if the chunk is stored as part of the original file or individually. If it's stored as part
	// of the file, then the location of the chunk in that file needs to be located before reading.
	StoredAsFile bool
	FilePath     string
}

// DataStore represents a collection of files.
type DataStore struct {
	Files []*File
}

// NewFile creates a new File and computes its chunks using the provided chunk size.
// reader is an IO reader that provides access to the underlying file contents.
// path is the expected filepath of the file, used for directory tree purposes.
// chunkSize is the number of bytes that each chunk should be at maximum.
func NewFile(reader FileIOReader, name string, chunkSize int) (*File, error) {
	// validate arguments
	if chunkSize <= 0 {
		return nil, errors.New("Chunk size must be a positive integer.")
	}

	// generate each chunk
	var chunks = Chunks{
		ChunkSize: chunkSize,
	}

	i := 0
	allContents := make([]byte, 0)
	buffer := make([]byte, chunkSize)
	stop := false
	for !stop {
		offset := int64(chunkSize * i)
		numRead, err := reader.ReadAt(buffer, offset)
		if err == io.EOF && numRead == 0 {
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
			ID:             chunkID,
			SequenceNumber: i,
			ChunkOffset:    offset,
			ContentSize:    numRead,
		}
		allContents = append(allContents, buffer[:chunk.ContentSize]...)
		chunks.Chunks = append(chunks.Chunks, chunk)
		chunks.NumChunks++
		i++
	}

	// Compute file ID by hashing all of the chunk IDs.
	id := generateFileID(chunks.Chunks)

	// compute extra information
	fileSize := chunks.ComputeFileSize()
	// Alternatively:
	// numChunks*chunkSize ~= fileSize
	// numChunks ~= ceil(fileSize/chunkSize)

	return &File{
		ID:     FileID(id),
		reader: reader,
		Name:   name,
		Size:   FileSize(fileSize),
		Chunks: chunks,
	}, nil
}

// Contains returns whether the datastore contains the specified file.
func (ds *DataStore) Contains(file *File) bool {
	for _, f := range ds.Files {
		if file.Name == f.Name {
			return true
		}
	}
	return false
}

func (ds *DataStore) ContainsName(name string) bool {
	for _, f := range ds.Files {
		if name == f.Name {
			return true
		}
	}
	return false
}

// Add appends a file to the datastore.
func (ds *DataStore) Add(file *File) {
	ds.Files = append(ds.Files, file)
}

func generateFileID(chunks []Chunk) string {
	buffer := make([]byte, 0)
	for _, c := range chunks {
		buffer = append(buffer, c.ID...)
	}
	return utils.HashFile(buffer)
}
