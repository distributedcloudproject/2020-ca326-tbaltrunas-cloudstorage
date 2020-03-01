package network

import (
	"cloud/datastore"
	"cloud/utils"
	"encoding/gob"
	"errors"
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
// TODO: might want to do the actual distribution here, so that the file gets saved with this call.
// TODO: Use reader instead of localPath.
func (c *cloud) AddFile(file *datastore.File, cloudPath string, localPath string) error {
	utils.GetLogger().Printf("[INFO] Sending AddFile request for file: %v, on node: %v.", file, c.MyNode().ID)
	_, err := c.SendMessageToMe(AddFileMsg, file, cloudPath)
	c.SendMessageAllOthers(AddFileMsg, file, cloudPath)
	utils.GetLogger().Printf("[DEBUG] Completed AddFile request for file: %v, on node: %v.", file, c.MyNode().ID)

	// TODO: move this to another function.
	for _, chunk := range file.Chunks.Chunks {
		c.DistributeChunk(datastore.ChunkStore{
			Chunk:        chunk,
			FileID:       file.ID,
			StoredAsFile: true,
			FilePath:     localPath,
		})
	}
	return err
}

func (r request) OnAddFileRequest(file *datastore.File, filepath string) error {
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

func (r request) OnUpdateFileRequest(file *datastore.File, filepath string) error {
	utils.GetLogger().Printf("[INFO] received UpdateFile request for file: %v from: %v.", filepath, r.FromNode.ID)

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
	oldFile := folder.Files.Files[found]
	folder.Files.Files[found] = file

	go func() {
		// Compute the difference in chunks.
		mb := make(map[string]struct{}, len(oldFile.Chunks.Chunks))
		for _, x := range oldFile.Chunks.Chunks {
			mb[string(x.ID)] = struct{}{}
		}
		var newChunks []datastore.Chunk
		var oldChunks []string
		for _, x := range file.Chunks.Chunks {
			if _, found := mb[string(x.ID)]; !found {
				newChunks = append(newChunks, x)
			} else {
				delete(mb, string(x.ID))
			}
		}
		for k := range mb {
			oldChunks = append(oldChunks, k)
		}
		// Delete old chunks.
		for _, chunk := range oldChunks {
			c.deleteStoredFileChunk(oldFile.ID, datastore.ChunkID(chunk))
		}
		// Download new chunks from the node.
		for _, chunk := range newChunks {
			res, err := r.FromNode.client.SendMessage(GetChunkMsg, string(chunk.ID))
			if err == nil {
				content := res[0].([]byte)
				c.storeChunk(file.ID, chunk, content)
			}
		}

		for _, fs := range c.fileSyncs {
			if fs.cloudPath == filepath {
				if c.watcher != nil {
					c.watcher.Remove(fs.localPath)
				}
				f, err := os.Create(fs.localPath)
				if err == nil {
					c.DownloadFile(*file, f)
				}
				f.Close()
				if c.watcher != nil {
					c.watcher.Add(fs.localPath)
				}
			}
		}
	}()
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

	var errs error
	for _, chunk := range folder.Files.Files[found].Chunks.Chunks {
		err = c.deleteStoredFileChunk(folder.Files.Files[found].ID, chunk.ID)
		if err != nil {
			if errs == nil {
				errs = errors.New("")
			}
			errs = errors.New(errs.Error() + err.Error() + ";")
		}
	}
	if errs != nil {
		utils.GetLogger().Printf("[WARNING] Failed to Delete chunks for %v: %v", filepath, errs)
	}
	folder.Files.Files = append(folder.Files.Files[:found], folder.Files.Files[found+1:]...)
	return errs
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
	return nil
}

type SaveChunkRequest struct {
	FileID datastore.FileID
	Chunk  datastore.Chunk // chunk metadata

	Contents []byte // chunk bytes
}

// SaveChunk persistently stores the chunkNum chunk on the node, using metadata from the file the chunk belongs to.
func (n *cloudNode) SaveChunk(fileID datastore.FileID, chunk datastore.Chunk, contents []byte) error {
	utils.GetLogger().Printf("[INFO] Sending SaveChunk request for file: %v, chunk number: %d, on node: %v.",
		fileID, chunk.SequenceNumber, n.ID)
	_, err := n.client.SendMessage(SaveChunkMsg, SaveChunkRequest{
		FileID:   fileID,
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
	if err := r.Cloud.storeChunk(sr.FileID, sr.Chunk, sr.Contents); err != nil {
		return err
	}

	utils.GetLogger().Printf("[DEBUG] Finished saving chunk.")

	err := r.Cloud.updateChunkNodes(sr.Chunk.ID, r.Cloud.MyNode().ID)
	return err
}

func (c *cloud) GetChunk(chunkID datastore.ChunkID) (content []byte, err error) {
	c.networkMutex.RLock()
	nodes := c.network.ChunkNodes[chunkID]
	c.networkMutex.RUnlock()

	for _, n := range nodes {
		cnode := c.GetCloudNode(n)
		if cnode != nil {
			res, err := cnode.client.SendMessage(GetChunkMsg, chunkID)
			if err == nil {
				return res[0].([]byte), nil
			}
		}
	}
	return nil, errors.New("could not download chunk")
}

func (r request) OnGetChunkRequest(chunkID datastore.ChunkID) (content []byte, err error) {
	c := r.Cloud

	c.chunkStorageMutex.RLock()
	storage := c.chunkStorage[string(chunkID)]
	c.chunkStorageMutex.RUnlock()

	if storage == nil || len(storage) == 0 {
		return nil, errors.New("chunk is not stored")
	}
	return c.readChunk(storage[0])
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
