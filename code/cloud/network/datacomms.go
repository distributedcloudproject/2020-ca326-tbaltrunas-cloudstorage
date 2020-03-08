package network

import (
	"cloud/datastore"
	"cloud/utils"
	"encoding/gob"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
)

const (
	CreateDirectoryMsg = "CreateDirectory"
	DeleteDirectoryMsg = "DeleteDirectory"

	AddFileMsg    = "AddFile"
	UpdateFileMsg = "UpdateFile"
	DeleteFileMsg = "DeleteFile"
	MoveFileMsg   = "MoveFile"

	SaveChunkMsg        = "SaveChunk"
	GetChunkMsg         = "GetChunk"
	updateChunkNodesMsg = "updateChunkNodes"

	LockFileMsg   = "LockFile"
	UnlockFileMsg = "UnlockFile"
)

func init() {
	gob.Register(&datastore.File{})
	gob.Register(SaveChunkRequest{})
	gob.Register(datastore.ChunkID(""))

	handlers = append(handlers, createDataStoreRequestHandler)
}

func (c *cloud) CreateDirectory(folderPath string) error {
	res := c.SendMessageAll(CreateDirectoryMsg, folderPath)
	for _, r := range res {
		if r.Error != nil {
			return r.Error
		}
	}
	return nil
}

func (r request) OnCreateDirectory(folderPath string) error {
	r.Cloud.networkMutex.RLock()
	defer r.Cloud.networkMutex.RUnlock()
	// GetFolder will create the folder if one doesn't exist.
	_, err := r.Cloud.network.GetFolder(folderPath)
	return err
}

// DeleteDirectory deletes the provided folder. The directory must be empty for the delete to be done.
func (c *cloud) DeleteDirectory(folderPath string) error {
	res := c.SendMessageAll(DeleteDirectoryMsg, folderPath)
	for _, r := range res {
		if r.Error != nil {
			return r.Error
		}
	}
	return nil
}

func (r request) OnDeleteDirectory(folderPath string) error {
	r.Cloud.networkMutex.RLock()
	defer r.Cloud.networkMutex.RUnlock()
	// GetFolder will create the folder if one doesn't exist.
	folder, err := r.Cloud.network.GetFolder(folderPath)
	if err != nil {
		return err
	}
	if len(folder.Files.Files) != 0 {
		return errors.New("directory is not empty")
	}

	splitFolders := strings.Split(folderPath, "/")
	baseFolder := strings.Join(splitFolders[:len(splitFolders)-1], "/")
	targetFolder := splitFolders[len(splitFolders)-1]
	var networkFolder *NetworkFolder

	if len(splitFolders) == 1 {
		networkFolder = r.Cloud.network.RootFolder
	} else {
		networkFolder, err = r.Cloud.network.GetFolder(baseFolder)
		if err != nil {
			return err
		}
	}

	for i := range networkFolder.SubFolders {
		if networkFolder.SubFolders[i].Name == targetFolder {
			networkFolder.SubFolders = append(networkFolder.SubFolders[:i], networkFolder.SubFolders[i+1:]...)
			return nil
		}
	}

	return errors.New("directory not found")
}

// AddFile adds a file to the Network. It distributes the file automatically.
// TODO: Use reader instead of LocalPath.
func (c *cloud) AddFile(file *datastore.File, cloudPath string, localPath string) error {
	var err error
	fs := c.FileStore(cloudPath)
	if fs == nil {
		fs, err = datastore.PartialFileStoreFromFile(file, localPath, c.config.FileStorageDir)
		if err != nil {
			return err
		}
		c.fileStorage[cloudPath] = fs
	}
	utils.GetLogger().Printf("[INFO] Sending AddFile request for file: %v, on node: %v.", file, c.MyNode().ID)
	_, err = c.SendMessageToMe(AddFileMsg, file, cloudPath)
	c.SendMessageAllOthers(AddFileMsg, file, cloudPath)
	utils.GetLogger().Printf("[DEBUG] Completed AddFile request for file: %v, on node: %v.", file, c.MyNode().ID)

	// TODO: move this to another function.

	fs = c.FileStore(cloudPath)
	if fs != nil {
		for _, chunk := range file.Chunks.Chunks {
			c.DistributeChunk(cloudPath, fs, chunk.ID)
		}
	}
	return err
}

func (c *cloud) AddFileMetadata(file *datastore.File, cloudPath string) error {
	utils.GetLogger().Printf("[INFO] Sending AddFile request for file: %v, on node: %v.", file, c.MyNode().ID)
	_, err := c.SendMessageToMe(AddFileMsg, file, cloudPath)
	if err != nil {
		return err
	}
	res := c.SendMessageAllOthers(AddFileMsg, file, cloudPath)
	for _, r := range res {
		if r.Error != nil {
			err = r.Error
		}
	}
	utils.GetLogger().Printf("[DEBUG] Completed AddFile request for file: %v, on node: %v.", file, c.MyNode().ID)

	// TODO: move this to another function.

	return err
}

func (c *cloud) AddFileSync(file *datastore.File, cloudPath string, localPath string) error {
	var err error
	fs := c.FileStore(cloudPath)
	if fs == nil {
		fs = &datastore.FullFileStore{
			BaseFileStore: datastore.BaseFileStore{
				FileID: file.ID,
				Chunks: file.Chunks.Chunks,
			},
			FilePath: localPath,
		}
		c.fileStorage[cloudPath] = fs
	}
	utils.GetLogger().Printf("[INFO] Sending AddFile request for file: %v, on node: %v.", file, c.MyNode().ID)
	_, err = c.SendMessageToMe(AddFileMsg, file, cloudPath)
	c.SendMessageAllOthers(AddFileMsg, file, cloudPath)
	utils.GetLogger().Printf("[DEBUG] Completed AddFile request for file: %v, on node: %v.", file, c.MyNode().ID)

	// TODO: move this to another function.

	fs = c.FileStore(cloudPath)
	if fs != nil {
		for _, chunk := range file.Chunks.Chunks {
			c.DistributeChunk(cloudPath, fs, chunk.ID)
		}
	}
	return err
}

func (c *cloud) AddFileInPlace(file *datastore.File, cloudPath string, localPath string) error {
	var err error
	fs := c.FileStore(cloudPath)
	if fs == nil {
		fs = &datastore.FullFileStore{
			BaseFileStore: datastore.BaseFileStore{
				FileID: file.ID,
				Chunks: file.Chunks.Chunks,
			},
			FilePath: localPath,
		}
		c.fileStorage[cloudPath] = fs
	}
	utils.GetLogger().Printf("[INFO] Sending AddFile request for file: %v, on node: %v.", file, c.MyNode().ID)
	_, err = c.SendMessageToMe(AddFileMsg, file, cloudPath)
	c.SendMessageAllOthers(AddFileMsg, file, cloudPath)
	utils.GetLogger().Printf("[DEBUG] Completed AddFile request for file: %v, on node: %v.", file, c.MyNode().ID)
	return err
}

func (r request) OnAddFileRequest(file *datastore.File, filepath string) error {
	filepath = CleanNetworkPath(filepath)
	utils.GetLogger().Printf("[INFO] Node: %v, received AddFile request for file: %v.", r.Cloud.MyNode().ID, file)

	c := r.Cloud
	c.networkMutex.RLock()
	folder, err := c.network.GetFolder(path.Dir(filepath))
	if err != nil {
		c.networkMutex.RUnlock()
		return err
	}
	ok := folder.Files.Contains(file)
	c.networkMutex.RUnlock()
	if ok {
		return nil
	}

	c.networkMutex.Lock()
	folder.Files.Add(file)
	c.networkMutex.Unlock()

	c.fileStorageMutex.Lock()
	storage := c.fileStorage[filepath]
	if storage == nil {
		if ok, fpath := c.isInFolderSync(filepath); ok {
			c.fileStorage[filepath] = &datastore.FullFileStore{
				BaseFileStore: datastore.BaseFileStore{
					FileID: file.ID,
					Chunks: file.Chunks.Chunks,
				},
				FilePath: fpath,
			}
			fil, err := os.Create(fpath)
			if err == nil {
				fil.Close()
			}
		} else {
			c.fileStorage[filepath] = &datastore.PartialFileStore{
				BaseFileStore: datastore.BaseFileStore{
					FileID: file.ID,
					Chunks: file.Chunks.Chunks,
				},
				FolderPath: c.config.FileStorageDir,
			}
		}
	}
	c.fileStorageMutex.Unlock()

	return nil
}

// UpdateFile updates a file on the network's data store.
// Does not update the actual chunks. File lock must be acquired for given path before.
func (c *cloud) UpdateFile(file *datastore.File, cloudPath string) error {
	_, err := c.SendMessageToMe(UpdateFileMsg, file, cloudPath)
	if err != nil {
		return err
	}
	res := c.SendMessageAllOthers(UpdateFileMsg, file, cloudPath)
	for _, r := range res {
		if r.Error != nil {
			return r.Error
		}
	}
	return nil
}

func (r request) OnUpdateFileRequest(file *datastore.File, cloudpath string) error {
	cloudpath = CleanNetworkPath(cloudpath)
	utils.GetLogger().Printf("[INFO] received UpdateFile request for file: %v from: %v.", cloudpath, r.FromNode.ID)

	c := r.Cloud

	r.Cloud.fileLockMutex.Lock()
	lockedBy, isLocked := r.Cloud.fileLocks[cloudpath]
	r.Cloud.fileLockMutex.Unlock()
	if !isLocked || lockedBy != r.FromNode.ID {
		return errors.New("node does not have the lock for the file acquired")
	}

	foldername, filename := path.Split(cloudpath)
	c.networkMutex.Lock()
	defer c.networkMutex.Unlock()
	folder, err := c.network.GetFolder(foldername)
	if err != nil {
		return err
	}

	// Check that there is a file with existing name.
	found := -1
	for i := range folder.Files.Files {
		if folder.Files.Files[i].Name == filename {
			found = i
		}
	}
	if found == -1 {
		return errors.New("file " + filename + " was not found")
	}
	folder.Files.Files[found] = file

	c.fileStorageMutex.Lock()
	defer c.fileStorageMutex.Unlock()

	if fileStore := c.fileStorage[cloudpath]; fileStore != nil {
		newChunks, _ := fileStore.SetChunks(file.Chunks.Chunks)
		if sf, ok := fileStore.(*datastore.SyncFileStore); ok {
			go func() {
				sf.StopWatching()
				for _, chunk := range newChunks {
					if r.FromNode.ID != c.MyNode().ID {
						res, err := r.FromNode.client.SendMessage(GetChunkMsg, cloudpath, chunk.ID)
						fmt.Println("Got", r.FromNode.ID, len(res[0].([]byte)), err)
						if err == nil {
							content := res[0].([]byte)
							sf.StoreChunk(chunk.ID, content)
						}
					}
					go c.updateChunkNodes(chunk.ID, r.Cloud.MyNode().ID)
				}
				sf.StartWatching()
			}()
		}
	}
	return nil
}

// DeleteFile deletes a file on the network's data store.
// Will delete stored chunks as well. File lock must be acquired for given path before.
func (c *cloud) DeleteFile(path string) error {
	_, err := c.SendMessageToMe(DeleteFileMsg, path)
	if err != nil {
		return err
	}
	res := c.SendMessageAllOthers(DeleteFileMsg, path)
	for _, r := range res {
		if r.Error != nil {
			return r.Error
		}
	}
	return nil
}

func (r request) OnDeleteFileRequest(filepath string) error {
	filepath = CleanNetworkPath(filepath)
	utils.GetLogger().Printf("[INFO] received DeleteFile request for file: %v from: %v.", filepath, r.FromNode.ID)

	c := r.Cloud

	r.Cloud.fileLockMutex.Lock()
	lockedBy, isLocked := r.Cloud.fileLocks[filepath]
	r.Cloud.fileLockMutex.Unlock()
	if !isLocked || lockedBy != r.FromNode.ID {
		return errors.New("node does not have the lock for the file acquired")
	}

	foldername, filename := path.Split(filepath)
	c.networkMutex.Lock()
	defer c.networkMutex.Unlock()
	folder, err := c.network.GetFolder(foldername)
	if err != nil {
		return err
	}

	// Check that there is a file with existing name.
	found := -1
	for i := range folder.Files.Files {
		if folder.Files.Files[i].Name == filename {
			found = i
		}
	}
	if found == -1 {
		return errors.New("file " + filename + " was not found")
	}

	folder.Files.Files = append(folder.Files.Files[:found], folder.Files.Files[found+1:]...)

	r.Cloud.fileStorageMutex.RLock()
	storage := r.Cloud.fileStorage[filepath]
	r.Cloud.fileStorageMutex.RUnlock()

	storage.DeleteAllContent()
	//storage.SetChunks([]datastore.Chunk{})

	return nil
}

// MoveFile moves a file from old path to new path.
// File lock must be acquired for old path and new path.
func (c *cloud) MoveFile(path string, newpath string) error {
	_, err := c.SendMessageToMe(MoveFileMsg, path, newpath)
	if err != nil {
		return err
	}
	res := c.SendMessageAllOthers(MoveFileMsg, path, newpath)
	for _, r := range res {
		if r.Error != nil {
			return r.Error
		}
	}
	return nil
}

func (r request) OnMoveFileRequest(filepath string, newfilepath string) error {
	filepath = CleanNetworkPath(filepath)
	newfilepath = CleanNetworkPath(newfilepath)
	utils.GetLogger().Printf("[INFO] received MoveFile request for file: %v from: %v.", filepath, r.FromNode.ID)

	c := r.Cloud

	r.Cloud.fileLockMutex.Lock()
	lockedBy, isLocked := r.Cloud.fileLocks[filepath]
	newLockedBy, newIsLocked := r.Cloud.fileLocks[newfilepath]
	r.Cloud.fileLockMutex.Unlock()
	if !isLocked || lockedBy != r.FromNode.ID {
		return errors.New("node does not have the lock for the file acquired")
	}
	if !newIsLocked || newLockedBy != r.FromNode.ID {
		return errors.New("node does not have the lock for the move-to file acquired")
	}

	folderName, filename := path.Split(filepath)
	c.networkMutex.Lock()
	defer c.networkMutex.Unlock()
	folder, err := c.network.GetFolder(folderName)

	if err != nil {
		return err
	}
	newFolderName, newFilename := path.Split(newfilepath)
	newFolder, err := c.network.GetFolder(newFolderName)
	if err != nil {
		return err
	}
	if newFolder.Files.ContainsName(newFilename) {
		return errors.New("file with that name already exists")
	}

	// Get current file and remove it.
	found := -1
	for i := range folder.Files.Files {
		if folder.Files.Files[i].Name == filename {
			found = i
		}
	}
	if found == -1 {
		return errors.New("file " + filename + " was not found")
	}
	file := folder.Files.Files[found]
	folder.Files.Files = append(folder.Files.Files[:found], folder.Files.Files[found+1:]...)
	file.Name = newFilename
	newFolder.Files.Add(file)

	c.fileStorageMutex.Lock()
	c.fileStorage[newfilepath] = c.fileStorage[filepath]
	delete(c.fileStorage, filepath)
	c.fileStorageMutex.Unlock()
	return nil
}

type SaveChunkRequest struct {
	FilePath string
	Chunk    datastore.Chunk // chunk metadata

	Contents []byte // chunk bytes
}

// SaveChunk persistently stores the chunkNum chunk on the node, using metadata from the file the chunk belongs to.
func (n *cloudNode) SaveChunk(filePath string, chunk datastore.Chunk, contents []byte) error {
	utils.GetLogger().Printf("[INFO] Sending SaveChunk request for file: %v, chunk number: %d, on node: %v.",
		filePath, chunk.SequenceNumber, n.ID)
	_, err := n.client.SendMessage(SaveChunkMsg, SaveChunkRequest{
		FilePath: filePath,
		Chunk:    chunk,
		Contents: contents,
	})
	return err
}

// OnSaveChunkRequest persistently stores a chunk given by its contents, as the given cloud path.
func (r request) OnSaveChunkRequest(sr SaveChunkRequest) error {
	utils.GetLogger().Printf("[INFO] Node: %v, received SaveChunk request.", r.Cloud.MyNode().ID)
	utils.GetLogger().Printf("[DEBUG] Got SaveChunkRequest chunk: %v.", sr.Chunk)

	// TODO: verify chunk ID

	utils.GetLogger().Println("[DEBUG] Created/verified existence of path directories.")
	r.Cloud.fileStorageMutex.RLock()
	storage := r.Cloud.fileStorage[sr.FilePath]
	r.Cloud.fileStorageMutex.RUnlock()
	if storage == nil {
		return errors.New("no storage found for file")
	}
	if err := storage.StoreChunk(sr.Chunk.ID, sr.Contents); err != nil {
		return err
	}
	utils.GetLogger().Printf("[DEBUG] Finished saving chunk.")

	benchmarkState := r.Cloud.BenchmarkState()
	benchmarkState.StorageSpaceUsed += uint64(len(sr.Contents))
	r.Cloud.SetBenchmarkState(benchmarkState)

	err := r.Cloud.updateChunkNodes(sr.Chunk.ID, r.Cloud.MyNode().ID)
	return err
}

func (c *cloud) GetChunk(filePath string, chunkID datastore.ChunkID) (content []byte, err error) {
	filePath = CleanNetworkPath(filePath)
	c.networkMutex.RLock()
	nodes := c.network.ChunkNodes[chunkID]
	c.networkMutex.RUnlock()

	for _, n := range nodes {
		cnode := c.GetCloudNode(n)
		if cnode != nil {
			res, err := cnode.client.SendMessage(GetChunkMsg, filePath, chunkID)
			if err == nil {
				return res[0].([]byte), nil
			}
			return nil, err
		}
	}
	return nil, errors.New("could not download chunk")
}

func (r request) OnGetChunkRequest(filePath string, chunkID datastore.ChunkID) (content []byte, err error) {
	c := r.Cloud

	c.fileStorageMutex.RLock()
	storage := c.fileStorage[filePath]
	c.fileStorageMutex.RUnlock()

	if storage == nil {
		return nil, errors.New("file is not stored")
	}
	return storage.ReadChunk(chunkID)
}

// updateChunkNodes updates the node's ChunkNodes data structure.
// It maps the chunkID key and appends the nodeID value.
// updateChunkNodes recursively sends out the update to other nodes.
func (c *cloud) updateChunkNodes(chunkID datastore.ChunkID, nodeID string) error {
	utils.GetLogger().Printf("[INFO] Sending updateChunkNodes request.")
	_, err := c.SendMessageToMe(updateChunkNodesMsg, chunkID, nodeID)
	if err != nil {
		// FIXME: a way to propagate errors returned from requests, i.e. take the place of communication.go errors
		utils.GetLogger().Printf("[ERROR] %v.", err)
	}
	c.SendMessageAllOthers(updateChunkNodesMsg, chunkID, nodeID)
	return err
}

func (r request) onUpdateChunkNodes(chunkID datastore.ChunkID, nodeID string) {
	utils.GetLogger().Printf("[INFO] Node: %v, received onUpdateChunkNodes request.", r.FromNode.ID)
	utils.GetLogger().Printf("[DEBUG] Received request at client: %v.", &r.FromNode.client)
	utils.GetLogger().Printf("[DEBUG] Updating ChunkNodes with ChunkID: %v, NodeID: %v.",
		chunkID, nodeID)

	c := r.Cloud
	c.networkMutex.Lock()
	defer c.networkMutex.Unlock()
	chunkNodes, ok := c.network.ChunkNodes[chunkID]
	if ok {
		utils.GetLogger().Printf("[DEBUG] Got list of nodes for key (%v): %v.", chunkID, chunkNodes)
		for _, nID := range chunkNodes {
			if nID == nodeID {
				// node already added
				// pre-emptive exit.
				utils.GetLogger().Println("[DEBUG] ChunkNodes already contains the needed key-value.")
				return
			}
		}
	}

	// update our own chunk location data structure
	utils.GetLogger().Printf("[DEBUG] Updating ChunkNodes: %v.", c.network.ChunkNodes)
	if !ok {
		// key not present
		utils.GetLogger().Printf("[DEBUG] Creating a new list for the key: %v, in ChunkNodes.", chunkID)
		chunkNodes = []string{nodeID}
	} else {
		// key present and has other nodes
		utils.GetLogger().Printf("[DEBUG] Appending to the list for the key: %v, in ChunkNodes.", chunkID)
		chunkNodes = append(chunkNodes, nodeID)
	}
	c.network.ChunkNodes[chunkID] = chunkNodes
	utils.GetLogger().Printf("[DEBUG] Finished updating ChunkNodes: %v.", c.network.ChunkNodes)
}

func (c *cloud) LockFile(path string) bool {
	path = CleanNetworkPath(path)
	// Check that we can lock the file on our client first.
	_, err := c.SendMessageToMe(LockFileMsg, path)
	if err != nil {
		// If we cannot acquire the lock, attempt to unlock for good measure.
		go c.UnlockFile(path)
		return false
	}
	// If we acquired the lock, tell all other nodes to lock the file.
	responses := c.SendMessageAllOthers(LockFileMsg, path)
	for _, res := range responses {
		// If any node couldn't lock the file, abandon.
		if res.Error != nil {
			go c.UnlockFile(path)
			return false
		}
	}

	// We have the file lock acquired.
	return true
}

func (r request) OnLockFileRequest(path string) error {
	r.Cloud.fileLockMutex.Lock()
	defer r.Cloud.fileLockMutex.Unlock()
	lockedBy, isLocked := r.Cloud.fileLocks[path]

	// If the file is already locked by the node requesting, return success.
	if isLocked && lockedBy == r.FromNode.ID {
		return nil
	}

	if isLocked {
		return errors.New("file locked by: " + lockedBy)
	}
	r.Cloud.fileLocks[path] = r.FromNode.ID
	return nil
}

func (c *cloud) UnlockFile(path string) {
	path = CleanNetworkPath(path)
	c.SendMessageAll(UnlockFileMsg, path)
}

func (r request) OnUnlockFileRequest(path string) error {
	r.Cloud.fileLockMutex.Lock()
	defer r.Cloud.fileLockMutex.Unlock()
	lockedBy, isLocked := r.Cloud.fileLocks[path]

	if !isLocked {
		return nil
	}

	if lockedBy != r.FromNode.ID {
		return errors.New("only lock owner may unlock the file")
	}

	delete(r.Cloud.fileLocks, path)
	return nil
}

func createDataStoreRequestHandler(node *cloudNode, cloud *cloud) func(string) interface{} {
	utils.GetLogger().Printf("[INFO] Creating a datastore request handler for node: %v, and cloud: %v.", node, cloud)
	r := request{
		Cloud:    cloud,
		FromNode: node,
	}
	utils.GetLogger().Printf("[DEBUG] Initialised request with fields: %v.", r)

	return func(message string) interface{} {
		switch message {
		case AddFileMsg:
			return r.OnAddFileRequest
		case SaveChunkMsg:
			return r.OnSaveChunkRequest
		case GetChunkMsg:
			return r.OnGetChunkRequest
		case updateChunkNodesMsg:
			return r.onUpdateChunkNodes
		case LockFileMsg:
			return r.OnLockFileRequest
		case UnlockFileMsg:
			return r.OnUnlockFileRequest
		case UpdateFileMsg:
			return r.OnUpdateFileRequest
		case MoveFileMsg:
			return r.OnMoveFileRequest
		case DeleteFileMsg:
			return r.OnDeleteFileRequest
		case CreateDirectoryMsg:
			return r.OnCreateDirectory
		case DeleteDirectoryMsg:
			return r.OnDeleteDirectory
		}
		return nil
	}
}
