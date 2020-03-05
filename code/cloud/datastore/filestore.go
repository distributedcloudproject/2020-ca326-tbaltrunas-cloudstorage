package datastore

import (
	"cloud/utils"
	"errors"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

type FileStore interface {
	HasChunk(chunkID ChunkID) bool
	Chunk(chunkID ChunkID) (Chunk, bool)
	SetChunks(chunks []Chunk) (newChunks []Chunk, oldChunks []Chunk)
	StoreChunk(chunkID ChunkID, content []byte) error
	ReadChunk(chunkID ChunkID) ([]byte, error)
	DeleteAllContent() error
}

type BaseFileStore struct {
	FileID FileID
	Chunks []Chunk
}

func (f *BaseFileStore) HasChunk(chunkID ChunkID) bool {
	for i := range f.Chunks {
		if f.Chunks[i].ID == chunkID {
			return true
		}
	}
	return false
}

func (f *BaseFileStore) Chunk(chunkID ChunkID) (Chunk, bool) {
	for i := range f.Chunks {
		if f.Chunks[i].ID == chunkID {
			return f.Chunks[i], true
		}
	}
	return Chunk{}, false
}

func (f *BaseFileStore) SetChunks(chunks []Chunk) (newChunks []Chunk, oldChunks []Chunk) {
	max := len(f.Chunks)
	// There's less chunks.
	if len(chunks) < max {
		oldChunks = append(oldChunks, f.Chunks[max:]...)
		max = len(chunks)
		f.Chunks = f.Chunks[:max]
	}

	for i := 0; i < max; i++ {
		// If the chunk ID is not the same(content does not match), set as new chunk.
		if f.Chunks[i].ID != chunks[i].ID {
			oldChunks = append(oldChunks, f.Chunks[i])
			newChunks = append(newChunks, chunks[i])
			f.Chunks[i] = chunks[i]
		}
	}

	// There's more chunks.
	if len(chunks) > max {
		f.Chunks = append(f.Chunks, chunks[max:]...)
		newChunks = append(newChunks, chunks[max:]...)
	}

	return
}

type FullFileStore struct {
	BaseFileStore
	// Path to the file.
	FilePath string

	mutex sync.RWMutex
}

func (f *FullFileStore) DeleteAllContent() error {
	return os.Remove(f.FilePath)
}

func (f *FullFileStore) SetChunks(chunks []Chunk) ([]Chunk, []Chunk) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	newChunks, oldChunks := f.BaseFileStore.SetChunks(chunks)

	file, err := os.OpenFile(f.FilePath, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		utils.GetLogger().Printf("[ERROR] Opening file %v: %v", f.FilePath, err)
		return newChunks, oldChunks
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		utils.GetLogger().Printf("[ERROR] File Stat %v: %v", f.FilePath, err)
		return newChunks, oldChunks
	}

	var newSize int64
	for i := range f.Chunks {
		newSize += int64(f.Chunks[i].ContentSize)
	}

	if newSize < info.Size() {
		file.Truncate(newSize)
	}

	return newChunks, oldChunks
}

func (f *FullFileStore) ReadChunk(chunkID ChunkID) ([]byte, error) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	c, found := f.Chunk(chunkID)
	if !found {
		return nil, errors.New("chunk does not belong to the file")
	}

	file, err := os.OpenFile(f.FilePath, os.O_RDONLY, 0666)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error opening file %v: %v", f.FilePath, err))
	}
	defer file.Close()

	pos, err := file.Seek(c.ChunkOffset, io.SeekStart)
	if err != nil {
		return nil, err
	}
	if pos != c.ChunkOffset {
		return nil, errors.New(fmt.Sprintf("mismatched position %d with offset: %d", pos, c.ChunkOffset))
	}

	content := make([]byte, c.ContentSize)
	read := 0
	for read < len(content) {
		r, err := file.Read(content)
		if err != nil {
			return nil, err
		}
		read += r
	}
	return content, nil
}

func (f *FullFileStore) StoreChunk(chunkID ChunkID, content []byte) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	c, found := f.Chunk(chunkID)
	if !found {
		return errors.New("chunk does not belong to the file")
	}

	file, err := os.OpenFile(f.FilePath, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return errors.New(fmt.Sprintf("error opening file %v: %v", f.FilePath, err))
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	// If the file is smaller than our chunk offset, append empty content to fill the space.
	if info.Size() < c.ChunkOffset {
		_, err = file.Seek(0, io.SeekEnd)
		if err != nil {
			return err
		}

		toWrite := c.ChunkOffset - info.Size()
		for toWrite > 0 {
			buffer := make([]byte, toWrite)
			w, err := file.Write(buffer)
			if err != nil {
				return err
			}
			toWrite -= int64(w)
		}
	}
	_, err = file.Seek(c.ChunkOffset, io.SeekStart)
	if err != nil {
		return err
	}
	written := 0
	for written < len(content) {
		n, err := file.Write(content[written:])
		if err != nil {
			return err
		}
		written += n
	}
	return nil
}

type SyncFileStore struct {
	FullFileStore

	Watcher   *fsnotify.Watcher
	inWatcher bool
}

func (f *SyncFileStore) StartWatching() {
	if f.Watcher != nil {
		f.Watcher.Add(f.FilePath)
		f.inWatcher = true
	}
}

func (f *SyncFileStore) StopWatching() {
	if f.Watcher != nil {
		f.Watcher.Remove(f.FilePath)
		f.inWatcher = false
	}
}

func (f *SyncFileStore) StoreChunk(chunkID ChunkID, content []byte) error {
	if f.inWatcher && f.Watcher != nil {
		f.Watcher.Remove(f.FilePath)
		defer f.Watcher.Add(f.FilePath)
	}
	return f.FullFileStore.StoreChunk(chunkID, content)
}

type PartialFileStore struct {
	BaseFileStore
	// Path to the folder that will store the chunks.
	FolderPath string
}

func PartialFileStoreFromFile(file *File, filepath string, folderStore string) (*PartialFileStore, error) {
	wholeFile := &FullFileStore{
		BaseFileStore: BaseFileStore{
			FileID: file.ID,
			Chunks: file.Chunks.Chunks,
		},
		FilePath: filepath,
	}
	f := &PartialFileStore{
		BaseFileStore: BaseFileStore{
			FileID: file.ID,
			Chunks: file.Chunks.Chunks,
		},
		FolderPath: folderStore,
	}
	for _, chunk := range file.Chunks.Chunks {
		b, err := wholeFile.ReadChunk(chunk.ID)
		if err != nil {
			return nil, err
		}
		f.StoreChunk(chunk.ID, b)
	}
	return f, nil
}

func (f *PartialFileStore) DeleteAllContent() error {
	var err error
	for _, c := range f.Chunks {
		path := filepath.Join(f.FolderPath, fmt.Sprintf("%s.%d", f.FileID, c.SequenceNumber))
		err2 := os.Remove(path)
		if err2 != nil {
			err = err2
		}
	}
	return err
}

func (f *PartialFileStore) SetChunks(chunks []Chunk) ([]Chunk, []Chunk) {
	newChunks, oldChunks := f.BaseFileStore.SetChunks(chunks)

	// Delete old chunks.
	for _, c := range oldChunks {
		path := filepath.Join(f.FolderPath, fmt.Sprintf("%s.%d", f.FileID, c.SequenceNumber))
		err := os.Remove(path)
		if err != nil {
			utils.GetLogger().Printf("[ERROR] Removing %v: %v", path, err)
		}
	}

	return newChunks, oldChunks
}

func (f *PartialFileStore) ReadChunk(chunkID ChunkID) ([]byte, error) {
	c, found := f.Chunk(chunkID)
	if !found {
		return nil, errors.New("chunk does not belong to the file")
	}

	path := filepath.Join(f.FolderPath, fmt.Sprintf("%s.%d", f.FileID, c.SequenceNumber))
	content, err := ioutil.ReadFile(path)
	return content, err
}

func (f *PartialFileStore) StoreChunk(chunkID ChunkID, content []byte) error {
	c, found := f.Chunk(chunkID)
	if !found {
		return errors.New("chunk does not belong to the file")
	}

	path := filepath.Join(f.FolderPath, fmt.Sprintf("%s.%d", f.FileID, c.SequenceNumber))
	return ioutil.WriteFile(path, content, 0666)
}
