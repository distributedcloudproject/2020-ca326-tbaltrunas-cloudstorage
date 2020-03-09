# datastore
--
    import "cloud/cloud/datastore"


## Usage

#### func  SaveChunk

```go
func SaveChunk(w io.Writer, buffer []byte) error
```
SaveChunk writes a bytes buffer through a writer, until the buffer is fully
written.

#### type BaseFileStore

```go
type BaseFileStore struct {
	FileID FileID
	Chunks []Chunk
}
```


#### func (*BaseFileStore) Chunk

```go
func (f *BaseFileStore) Chunk(chunkID ChunkID) (Chunk, bool)
```

#### func (*BaseFileStore) HasChunk

```go
func (f *BaseFileStore) HasChunk(chunkID ChunkID) bool
```

#### func (*BaseFileStore) SetChunks

```go
func (f *BaseFileStore) SetChunks(chunks []Chunk) (newChunks []Chunk, oldChunks []Chunk)
```

#### type Chunk

```go
type Chunk struct {
	ID             ChunkID // Unique ID of the chunk (hash value of the contents).
	SequenceNumber int     // Chunk sequence used to place the chunk in the correct position in the file.

	ContentSize uint64 // Number of bytes of actual content.
	ChunkOffset int64  // The offset of the chunk of the file. (e.g, if offset is 1024, the chunk is contained at

}
```

Chunk represents a "chunk" of a file, a sequential part of a file. Each chunk
has an ID and a sequence number.

#### type ChunkID

```go
type ChunkID string
```

ChunkID is a hash as a string of bytes.

#### func  ComputeChunkID

```go
func ComputeChunkID(buffer []byte) ChunkID
```
ComputeChunkID calculates the ID (hash) of a buffer of bytes (a chunk).

#### type ChunkStore

```go
type ChunkStore struct {
	Chunk  Chunk
	FileID FileID

	// StoredAsFile states if the chunk is stored as part of the original file or individually. If it's stored as part
	// of the file, then the location of the chunk in that file needs to be located before reading.
	StoredAsFile bool
	FilePath     string
	ContentSize  uint64 // In bytes, actual content size.
}
```

ChunkStore represents a locally stored Chunk.

#### type Chunks

```go
type Chunks struct {
	NumChunks int // Number of chunks that this file is split into.

	ChunkSize int // The maximum size of each chunk.

	Chunks []Chunk // List of chunks belonging to the file.
}
```


#### func (*Chunks) ComputeFileSize

```go
func (chunks *Chunks) ComputeFileSize() uint64
```
ComputeFileSize calculates the combined size of all chunks (the expected "file
size").

#### type DataStore

```go
type DataStore struct {
	Files []*File
}
```

DataStore represents a collection of files.

#### func (*DataStore) Add

```go
func (ds *DataStore) Add(file *File)
```
Add appends a file to the datastore.

#### func (*DataStore) Contains

```go
func (ds *DataStore) Contains(file *File) bool
```
Contains returns whether the datastore contains the specified file.

#### func (*DataStore) ContainsName

```go
func (ds *DataStore) ContainsName(name string) bool
```

#### func (*DataStore) GetChunkByID

```go
func (ds *DataStore) GetChunkByID(chunkID ChunkID) (*Chunk, *File)
```
GetChunkByID searches for the chunk with the given ID and the file the chunk
belongs to. Returns nil if the chunk can not be found.

#### type File

```go
type File struct {
	ID FileID // ID of the file (hash of chunk IDs).

	Name string // The name of the file.

	Path string // Path of the user's file.

	Size uint64 // File size in bytes.

	Chunks Chunks // List of the file's chunk ID's.
}
```

File represents a user's file stored on the cloud.

#### func  NewFile

```go
func NewFile(reader FileIOReader, name string, chunkSize int) (*File, error)
```
NewFile creates a new File and computes its chunks using the provided chunk
size. reader is an IO reader that provides access to the underlying file
contents. path is the expected filepath of the file, used for directory tree
purposes. chunkSize is the number of bytes that each chunk should be at maximum.

#### func (*File) GetChunk

```go
func (file *File) GetChunk(n int) ([]byte, int, error)
```
GetChunk reads the nth chunk in the file. Returns the contents as bytes, the
amount of actual bytes read, and error if any.

#### func (*File) GetChunkByID

```go
func (file *File) GetChunkByID(chunkID ChunkID) *Chunk
```
GetChunkByID returns a chunk belonging to the file by its ID. Returns nil if the
chunk can not be found.

#### func (*File) LoadChunk

```go
func (file *File) LoadChunk(r io.Reader) ([]byte, error)
```
LoadChunk reads a chunk from a reader.

#### type FileID

```go
type FileID string
```

FileID is a hash as a string of bytes.

#### type FileIOReader

```go
type FileIOReader io.ReaderAt
```

FileIOReader is the type used by the File reader.

#### type FileStore

```go
type FileStore interface {
	HasChunk(chunkID ChunkID) bool
	Chunk(chunkID ChunkID) (Chunk, bool)
	SetChunks(chunks []Chunk) (newChunks []Chunk, oldChunks []Chunk)
	StoreChunk(chunkID ChunkID, content []byte) error
	ReadChunk(chunkID ChunkID) ([]byte, error)
	DeleteAllContent() error
}
```


#### type FullFileStore

```go
type FullFileStore struct {
	BaseFileStore
	// Path to the file.
	FilePath string
}
```


#### func (*FullFileStore) DeleteAllContent

```go
func (f *FullFileStore) DeleteAllContent() error
```

#### func (*FullFileStore) ReadChunk

```go
func (f *FullFileStore) ReadChunk(chunkID ChunkID) ([]byte, error)
```

#### func (*FullFileStore) SetChunks

```go
func (f *FullFileStore) SetChunks(chunks []Chunk) ([]Chunk, []Chunk)
```

#### func (*FullFileStore) StoreChunk

```go
func (f *FullFileStore) StoreChunk(chunkID ChunkID, content []byte) error
```

#### type PartialFileStore

```go
type PartialFileStore struct {
	BaseFileStore
	// Path to the folder that will store the chunks.
	FolderPath string
}
```


#### func  PartialFileStoreFromFile

```go
func PartialFileStoreFromFile(file *File, filepath string, folderStore string) (*PartialFileStore, error)
```

#### func (*PartialFileStore) DeleteAllContent

```go
func (f *PartialFileStore) DeleteAllContent() error
```

#### func (*PartialFileStore) ReadChunk

```go
func (f *PartialFileStore) ReadChunk(chunkID ChunkID) ([]byte, error)
```

#### func (*PartialFileStore) SetChunks

```go
func (f *PartialFileStore) SetChunks(chunks []Chunk) ([]Chunk, []Chunk)
```

#### func (*PartialFileStore) StoreChunk

```go
func (f *PartialFileStore) StoreChunk(chunkID ChunkID, content []byte) error
```

#### type SyncFileStore

```go
type SyncFileStore struct {
	FullFileStore
	CloudPath string

	LastEdit time.Time
}
```


#### func (*SyncFileStore) StoreChunk

```go
func (f *SyncFileStore) StoreChunk(chunkID ChunkID, content []byte) error
```

#### func (*SyncFileStore) WatcherEvent

```go
func (f *SyncFileStore) WatcherEvent(event *fsnotify.Event, fa *File) (*File, error)
```
