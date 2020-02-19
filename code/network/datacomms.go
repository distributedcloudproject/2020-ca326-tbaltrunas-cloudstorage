package network

import (
	"cloud/datastore"
	"cloud/utils"
	"encoding/gob"
	"os"
	"path/filepath"
	"strconv"
)

const (
	AddFileMsg          = "AddFile"
	SaveChunkMsg        = "SaveChunk"
	updateChunkNodesMsg = "updateChunkNodes"
)

type SaveChunkRequest struct {
	// CloudPath is the filepath for the chunk on the cloud.
	// The path should be rooted at "/" (without drive letter on Windows).
	// The actual storage path will depedend on the node's configuration.
	CloudPath string

	Chunk datastore.Chunk // chunk metadata

	Contents []byte // chunk bytes
}

func init() {
	gob.Register(&datastore.File{})
	gob.Register(SaveChunkRequest{})
	gob.Register(datastore.ChunkID(""))

	handlers = append(handlers, createDataStoreRequestHandler)
}

// AddFile adds a file to the Network's datastore.
// It does not add the actual chunks.
// TODO: might want to do the actual distribution here, so that the file gets saved with this call.
func (c *cloud) AddFile(file *datastore.File) error {
	utils.GetLogger().Printf("[INFO] Sending AddFile request for file: %v, on node: %v.", file, c.MyNode().ID)
	_, err := c.SendMessageToMe(AddFileMsg, file)
	go c.SendMessageAllOthers(AddFileMsg, file)
	utils.GetLogger().Printf("[DEBUG] Completed AddFile request for file: %v, on node: %v.", file, c.MyNode().ID)
	return err
}

func (r request) OnAddFileRequest(file *datastore.File) error {
	utils.GetLogger().Printf("[INFO] Node: %v, received AddFile request for file: %v.", r.Cloud.MyNode().ID, file)

	c := r.Cloud
	c.networkMutex.RLock()
	ok := c.network.DataStore.Contains(file)
	c.networkMutex.RUnlock()
	if ok {
		return nil
	}

	c.networkMutex.Lock()
	c.network.DataStore.Add(file)
	c.networkMutex.Unlock()

	return nil
}

// SaveChunk persistently stores the chunkNum chunk on the node, using metadata from the file the chunk belongs to.
func (n *cloudNode) SaveChunk(file *datastore.File, chunkNum int) error {
	utils.GetLogger().Printf("[INFO] Sending SaveChunk request for file: %v, chunk number: %d, on node: %v.",
		file.Path, chunkNum, n.ID)
	chunk := file.Chunks.Chunks[chunkNum]
	contents, _, err := file.GetChunk(chunkNum)
	cloudPath := filepath.Base(file.Path) + "-" + strconv.Itoa(chunk.SequenceNumber)
	_, err = n.client.SendMessage(SaveChunkMsg, SaveChunkRequest{
		CloudPath: cloudPath,
		Chunk:     chunk,
		Contents:  contents,
	})
	return err
}

// OnSaveChunkRequest persistently stores a chunk given by its contents, as the given cloud path.
func (r request) OnSaveChunkRequest(sr SaveChunkRequest) error {
	utils.GetLogger().Printf("[INFO] Node: %v, received SaveChunk request.", r.Cloud.MyNode().ID)
	cloudPath := sr.CloudPath
	chunk := sr.Chunk
	contents := sr.Contents
	utils.GetLogger().Printf("[DEBUG] Got SaveChunkRequest with path: %v, chunk: %v.", cloudPath, chunk)

	// TODO: verify chunk ID
	// chunkID := chunk.ID

	// persistently store chunk
	c := r.Cloud
	c.Mutex.Lock()

	localPath := filepath.Join(c.config.FileStorageDir, cloudPath)
	utils.GetLogger().Printf("[DEBUG] Computed path where to store chunk: %s.", localPath)

	// TODO: maybe this should be done when setting up the node
	err := os.MkdirAll(filepath.Dir(localPath), os.ModeDir)
	if err != nil {
		c.Mutex.Unlock()
		return err
	}

	utils.GetLogger().Println("[DEBUG] Created/verified existence of path directories.")
	w, err := os.Create(localPath)
	if err != nil {
		c.Mutex.Unlock()
		return err
	}
	defer w.Close()

	utils.GetLogger().Printf("[DEBUG] Saving chunk to writer: %v.", w)
	err = datastore.SaveChunk(w, contents)
	c.Mutex.Unlock()
	if err != nil {
		return err
	}
	utils.GetLogger().Printf("[DEBUG] Finished saving chunk to writer: %v.", w)

	err = c.updateChunkNodes(chunk.ID, c.MyNode().ID)
	return err
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
	go c.SendMessageAllOthers(updateChunkNodesMsg, chunkID, nodeID)
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
		case updateChunkNodesMsg:
			return r.onUpdateChunkNodes
		}
		return nil
	}
}
