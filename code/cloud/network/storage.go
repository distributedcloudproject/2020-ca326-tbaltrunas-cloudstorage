package network

import (
	"cloud/datastore"
	"cloud/utils"
	"errors"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// DownloadFile downloads the file from the cloud.
func (c *cloud) DownloadFile(cloudPath string, w io.Writer) error {
	file, err := c.GetFile(cloudPath)
	if err != nil {
		return err
	}
	for _, chunk := range file.Chunks.Chunks {
		content, err := c.GetChunk(cloudPath, chunk.ID)
		if err != nil {
			return err
		}
		_, err = w.Write(content)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *cloud) isInFolderSync(cloudPath string) (ok bool, filePath string) {
	for i := range c.folderSyncs {
		if strings.HasPrefix(cloudPath, c.folderSyncs[i].CloudPath) {
			relPath := strings.TrimPrefix(cloudPath, c.folderSyncs[i].CloudPath)
			fpath := filepath.Join(c.folderSyncs[i].LocalPath, relPath)
			return true, fpath
		}
	}
	return false, ""
}

func (c *cloud) watcherEvent(event *fsnotify.Event) {
	for i := range c.folderSyncs {
		if strings.HasPrefix(event.Name, c.folderSyncs[i].LocalPath) {
			relativePath := strings.TrimPrefix(event.Name, c.folderSyncs[i].LocalPath)
			relativePath = filepath.ToSlash(relativePath)
			cloudPath := path.Join(c.folderSyncs[i].CloudPath, relativePath)
			cloudPath = CleanNetworkPath(cloudPath)
			if event.Op&fsnotify.Write == fsnotify.Write {
				utils.GetLogger().Println("[INFO] modified file:", event.Name, c.folderSyncs[i].LocalPath, relativePath, cloudPath)

				f, err := c.GetFile(cloudPath)
				if err != nil {
					continue
				}
				reader, err := os.Open(event.Name)
				if err != nil {
					continue
				}
				info, err := reader.Stat()
				if err != nil {
					continue
				}
				if info.ModTime() == c.folderSyncs[i].LastEditTime {
					continue
				}
				c.folderSyncs[i].LastEditTime = info.ModTime()

				f2, err := datastore.NewFile(reader, f.Name, f.Chunks.ChunkSize)
				if len(f2.Chunks.Chunks) == 0 {
					continue
				}
				reader.Close()
				if err == nil {
					if c.LockFile(cloudPath) {
						err := c.UpdateFile(f2, cloudPath)
						if err != nil {
							utils.GetLogger().Println("[ERROR] UpdateFile:", err)
						}
					}
					c.UnlockFile(cloudPath)
				}
			}
			if event.Op&fsnotify.Create == fsnotify.Create {
				utils.GetLogger().Println("[INFO] created file:", event.Name, c.folderSyncs[i].LocalPath, relativePath, cloudPath)
				reader, err := os.Open(event.Name)
				if err != nil {
					continue
				}
				f2, err := datastore.NewFile(reader, filepath.Base(event.Name), 4*1024*1024)
				if err != nil {
					utils.GetLogger().Println("[ERROR] getfile on created file:", err)
					continue
				}
				c.LockFile(cloudPath)
				err = c.AddFileSync(f2, cloudPath, event.Name)
				fmt.Println("Create: ", cloudPath, event.Name, err)
				c.UnlockFile(cloudPath)
			}
			if event.Op&fsnotify.Remove == fsnotify.Remove {
				utils.GetLogger().Println("[INFO] remove file:", event.Name, c.folderSyncs[i].LocalPath, relativePath, cloudPath)
				c.LockFile(cloudPath)
				c.DeleteFile(cloudPath)
				c.UnlockFile(cloudPath)
			}
		}
	}
}

func (c *cloud) createWatcher() error {
	if c.watcher == nil {
		var err error
		c.watcher, err = fsnotify.NewWatcher()
		if err != nil {
			return err
		}

		go func() {
			for {
				select {
				case event, ok := <-c.watcher.Events:
					if !ok {
						return
					}
					c.watcherEvent(&event)
					if event.Op&fsnotify.Write == fsnotify.Write {
						utils.GetLogger().Println("[INFO] modified file:", event.Name)

						for i, fs := range c.fileSyncs {
							if fs.LocalPath == event.Name {
								f, err := c.GetFile(fs.CloudPath)
								if err != nil {
									continue
								}
								reader, err := os.Open(fs.LocalPath)
								if err != nil {
									continue
								}
								info, err := reader.Stat()
								if err != nil {
									continue
								}
								if info.ModTime() == fs.LastEditTime {
									continue
								}
								c.fileSyncs[i].LastEditTime = info.ModTime()

								f2, err := datastore.NewFile(reader, f.Name, f.Chunks.ChunkSize)
								if len(f2.Chunks.Chunks) == 0 {
									continue
								}
								reader.Close()
								if err == nil {
									if c.LockFile(fs.CloudPath) {
										err := c.UpdateFile(f2, fs.CloudPath)
										if err != nil {
											utils.GetLogger().Println("[ERROR] UpdateFile:", err)
										}
									}
									c.UnlockFile(fs.CloudPath)
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
	return nil
}

// SyncFile syncs a cloud and local file. The local path will be uploaded to the cloud at provided path, and any changes
// to either will reflect upon the other, the sync is constant.
// If cloud path exists but local path does not, the file will be downloaded.
// If local path exists but cloud path does not, the file will be uploaded to the cloud.
// If both exist, the content MUST be the same.
// If neither exist, an error will be thrown.
// This function returns before download/upload is completed. TODO: another function to check status.
func (c *cloud) SyncFile(cloudPath string, localPath string) error {
	if c.watcher == nil {
		if err := c.createWatcher(); err != nil {
			return err
		}
	}
	cloudPath = CleanNetworkPath(cloudPath)
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
			c.fileStorageMutex.Lock()
			c.fileStorage[cloudPath] = &datastore.SyncFileStore{
				FullFileStore: datastore.FullFileStore{
					BaseFileStore: datastore.BaseFileStore{
						FileID: localFile.ID,
						Chunks: localFile.Chunks.Chunks,
					},
					FilePath: localPath,
				},
				Watcher: c.watcher,
			}
			c.fileStorageMutex.Unlock()
			c.AddFile(localFile, cloudPath, localPath)
			c.watcher.Add(localPath)
		}()
	} else if localFile == nil {
		osFile, err := os.Create(localPath)
		if err != nil {
			return err
		}
		go func(cloudPath string, osFile *os.File) {
			err = c.DownloadFile(cloudPath, osFile)
			osFile.Close()
			c.fileStorageMutex.Lock()
			c.fileStorage[cloudPath].DeleteAllContent()
			c.fileStorage[cloudPath] = &datastore.SyncFileStore{
				FullFileStore: datastore.FullFileStore{
					BaseFileStore: datastore.BaseFileStore{
						FileID: cloudFile.ID,
						Chunks: cloudFile.Chunks.Chunks,
					},
					FilePath: localPath,
				},
				Watcher: c.watcher,
			}
			c.fileStorageMutex.Unlock()
			c.watcher.Add(localPath)
		}(cloudPath, osFile)
	} else {
		c.fileStorageMutex.Lock()
		c.fileStorage[cloudPath] = &datastore.SyncFileStore{
			FullFileStore: datastore.FullFileStore{
				BaseFileStore: datastore.BaseFileStore{
					FileID: localFile.ID,
					Chunks: localFile.Chunks.Chunks,
				},
				FilePath: localPath,
			},
			Watcher: c.watcher,
		}
		c.fileStorageMutex.Unlock()
		c.watcher.Add(localPath)
	}
	c.fileSyncs = append(c.fileSyncs, fileSync{
		LocalPath: localPath,
		CloudPath: cloudPath,
	})
	return nil
}

func (c *cloud) SyncFolder(cloudPath string, localPath string) error {
	if c.watcher == nil {
		if err := c.createWatcher(); err != nil {
			return err
		}
	}

	cloudPath = CleanNetworkPath(cloudPath)
	if empty, err := IsDirEmpty(localPath); err != nil || !empty {
		if err != nil {
			return err
		}
		return errors.New("directory is not empty")
	}

	f, err := c.GetFolder(cloudPath)
	if err != nil {
		return err
	}

	c.folderSyncs = append(c.folderSyncs, fileSync{
		CloudPath: cloudPath,
		LocalPath: localPath,
	})

	var syncFolder func(f *NetworkFolder, folderpath string)
	syncFolder = func(folder *NetworkFolder, folderpath string) {
		localFolder := filepath.Join(localPath, folderpath)
		folderpath = CleanNetworkPath(folderpath)
		for _, sf := range folder.SubFolders {
			syncFolder(sf, folderpath+"/"+sf.Name)
			//localFolder = path.Join(LocalPath, folderpath, folder.Name)
			os.MkdirAll(filepath.Join(localPath, folderpath), 0666)
		}

		downloadsRem := 0
		for _, f := range folder.Files.Files {
			cloudFilePath := path.Join(cloudPath, folderpath, f.Name)
			cloudFilePath = CleanNetworkPath(cloudFilePath)
			localFilePath := path.Join(localPath, folderpath, f.Name)

			c.fileStorageMutex.Lock()
			if _, ok := c.fileStorage[cloudFilePath]; ok {
				c.fileStorage[cloudFilePath].DeleteAllContent()
			} else {
				utils.GetLogger().Println("[ERROR] File", cloudFilePath, "not found in fileStorage")
			}
			c.downloadManager.QueueDownload(cloudFilePath, localFilePath, func() {
				downloadsRem--
				if downloadsRem <= 0 {
					c.watcher.Add(localFolder)
				}
			})
			c.fileStorage[cloudFilePath] = &datastore.SyncFileStore{
				FullFileStore: datastore.FullFileStore{
					BaseFileStore: datastore.BaseFileStore{
						FileID: f.ID,
						Chunks: f.Chunks.Chunks,
					},
					FilePath: localFilePath,
				},
			}
			c.fileStorageMutex.Unlock()
		}
	}
	syncFolder(f, "/")
	return nil
}

func IsDirEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1) // Or f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err // Either not empty or error, suits both cases
}
