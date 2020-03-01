package network

import (
	"cloud/datastore"
	"cloud/utils"
	"errors"
	"github.com/fsnotify/fsnotify"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
)

func (c *cloud) readChunk(chunk datastore.ChunkStore) ([]byte, error) {
	if !chunk.StoredAsFile {
		path := filepath.Join(c.config.FileStorageDir, string(chunk.Chunk.ID))
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		content, err := ioutil.ReadAll(f)
		return content, err
	}

	path := chunk.FilePath
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	_, err = f.Seek(chunk.Chunk.ChunkOffset, 0)
	if err != nil {
		return nil, err
	}
	content := make([]byte, chunk.Chunk.ContentSize)
	totalRead := 0
	for totalRead < chunk.Chunk.ContentSize {
		n, err := f.Read(content[totalRead:])
		if err != nil {
			return nil, err
		}
		totalRead += n
	}
	return content, nil
}

func (c *cloud) storeChunk(fileID datastore.FileID, chunk datastore.Chunk, content []byte) error {
	chunkID := string(chunk.ID)
	path := filepath.Join(c.config.FileStorageDir, chunkID)

	cs := datastore.ChunkStore{
		Chunk:        chunk,
		FileID:       fileID,
		StoredAsFile: false,
		FilePath:     path,
	}

	// Check if the chunk ID is already stored.
	if _, err := os.Stat(path); err != nil {
		// Save the file.
		w, err := os.Create(path)
		if err != nil {
			return err
		}

		if err = datastore.SaveChunk(w, content); err != nil {
			return err
		}

		if err = w.Close(); err != nil {
			return err
		}
	}

	c.chunkStorageMutex.Lock()
	chunkStores, _ := c.chunkStorage[chunkID]
	chunkStores = append(chunkStores, cs)
	c.chunkStorage[chunkID] = chunkStores
	c.chunkStorageMutex.Unlock()
	return nil
}

func (c *cloud) deleteStoredFileChunk(FileID datastore.FileID, ChunkID datastore.ChunkID) error {
	chunkID := string(ChunkID)

	c.chunkStorageMutex.Lock()
	chunkStores := c.chunkStorage[chunkID]

	for i := 0; i < len(chunkStores); i++ {
		if chunkStores[i].FileID == FileID {
			chunkStores = append(chunkStores[:i], chunkStores[i+1:]...)
			i = i - 1
		}
	}
	c.chunkStorage[chunkID] = chunkStores
	c.chunkStorageMutex.Unlock()

	// If there's no file with that chunk, delete the corresponding file.
	if len(chunkStores) == 0 {
		path := filepath.Join(c.config.FileStorageDir, chunkID)
		err := os.Remove(path)
		if err != os.ErrNotExist {
			return err
		}
	}
	return nil
}

func (c *cloud) DownloadFile(file datastore.File, w io.Writer) error {
	for _, chunk := range file.Chunks.Chunks {
		content, err := c.readChunk(datastore.ChunkStore{
			Chunk:        chunk,
			FileID:       file.ID,
			StoredAsFile: false,
		})
		if err != nil {
			content, err = c.GetChunk(chunk.ID)
			if err != nil {
				return err
			}
		}
		_, err = w.Write(content)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *cloud) storeChunksOfFile(file *datastore.File, pathToFile string) {
	c.chunkStorageMutex.RLock()
	defer c.chunkStorageMutex.RUnlock()
	for _, chunk := range file.Chunks.Chunks {
		chunkID := string(chunk.ID)
		cs := datastore.ChunkStore{
			Chunk:        chunk,
			FileID:       file.ID,
			StoredAsFile: true,
			FilePath:     pathToFile,
		}
		c.chunkStorage[chunkID] = append(c.chunkStorage[chunkID], cs)
	}
}

// SyncFile syncs a cloud and local file. The local path will be uploaded to the cloud at provided path, and any changes
// to either will reflect upon the other, the sync is constant.
// If cloud path exists but local path does not, the file will be downloaded.
// If local path exists but cloud path does not, the file will be uploaded to the cloud.
// If both exist, the content MUST be the same.
// If neither exist, an error will be thrown.
// This function returns before download/upload is completed. TODO: another function to check status.
func (c *cloud) SyncFile(cloudPath string, localPath string) error {
	cloudPath = path.Clean(cloudPath)
	_, name := path.Split(cloudPath)
	var localFile *datastore.File

	if f, err := os.Open(localPath); err == nil {
		file, err := datastore.NewFile(f, name, 4*1024*1024)
		if err == nil {
			localFile = file
		}
		f.Close()
	}

	cloudFile, err := c.GetFile(cloudPath)
	if err != nil && localFile == nil {
		return errors.New("file does not exist on the cloud nor locally")
	}

	if cloudFile != nil && localFile != nil {
		if cloudFile.ID != localFile.ID {
			return errors.New("local file and cloud file are not the same")
		}
	}

	if cloudFile == nil {
		go func() {
			c.AddFile(localFile, cloudPath, localPath)
			c.storeChunksOfFile(localFile, localPath)
			c.watchLocalFile(localPath)
		}()
	} else if localFile == nil {
		osFile, err := os.Create(localPath)
		if err != nil {
			return err
		}
		go func(f datastore.File, osFile *os.File) {
			err = c.DownloadFile(f, osFile)
			osFile.Close()
			c.storeChunksOfFile(&f, localPath)
			c.watchLocalFile(localPath)
		}(*cloudFile, osFile)
	} else {
		c.storeChunksOfFile(localFile, localPath)
		c.watchLocalFile(localPath)
	}
	c.fileSyncs = append(c.fileSyncs, fileSync{
		localPath: localPath,
		cloudPath: cloudPath,
	})
	return nil
}

func (c *cloud) watchLocalFile(localPath string) error {
	if c.watcher == nil {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return err
		}
		c.watcher = watcher

		go func() {
			for {
				select {
				case event, ok := <-c.watcher.Events:
					if !ok {
						return
					}
					if event.Op&fsnotify.Write == fsnotify.Write {
						utils.GetLogger().Println("[INFO] modified file:", event.Name)

						for _, fs := range c.fileSyncs {
							if fs.localPath == event.Name {
								f, err := c.GetFile(fs.cloudPath)
								if err == nil {
									c.UpdateFile(f, fs.cloudPath)
								}
							}
						}
					}
				case err, ok := <-c.watcher.Errors:
					if !ok {
						return
					}
					utils.GetLogger().Println("[INFO] error:", err)
				}
			}
		}()
	}
	return c.watcher.Add(localPath)
}

func (c *cloud) RemoveSync(localPath string) {

}
