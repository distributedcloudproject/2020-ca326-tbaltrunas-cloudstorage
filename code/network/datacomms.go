package network

import (
	"cloud/utils"
	"cloud/datastore"
	"encoding/gob"
	"path/filepath"
	"strconv"
)
 	
const (
	AddFileMsg = "AddFile"
	SaveChunkMsg = "SaveChunk"
	updateChunkNodesMsg = "updateChunkNodes"
)

type SaveChunkRequest struct {
	// CloudPath is the filepath for the chunk on the cloud.
	// The path should be rooted at "/" (without drive letter on Windows).
	// The actual storage path will depedend on the node's configuration.
	CloudPath 	string

	Chunk 		datastore.Chunk  	  // chunk metadata

	Contents 	[]byte           	  // chunk bytes
}

func init() {
	gob.Register(&datastore.File{})
	gob.Register(SaveChunkRequest{})
	gob.Register(datastore.ChunkID(""))
}

func (n *Node) AddFile(file *datastore.File) error {
	utils.GetLogger().Printf("[INFO] Sending AddFile request for file: %v, on node: %v.", file, n)
	utils.GetLogger().Printf("[DEBUG] Node's client is: %v.", n.client)
	_, err := n.client.SendMessage(AddFileMsg, file)
	utils.GetLogger().Printf("[DEBUG] Completed AddFile request for file: %v, on node: %v.", file, n)
	return err
}

func (r request) OnAddFileRequest(file *datastore.File) error {
	utils.GetLogger().Printf("[INFO] Node: %v, received AddFile request for file: %v.", r.cloud.MyNode.ID, file)
	err := r.cloud.addFile(file)
	return err
}

func (n *Node) SaveChunk(file *datastore.File, chunkNum int) error {
	utils.GetLogger().Printf("[INFO] Sending SaveChunk request for file: %v, chunk number: %d, on node: %v.", 
							 file.Path, chunkNum, n.Name)
	chunk := file.Chunks.Chunks[chunkNum]
	contents, _, err := file.GetChunk(chunkNum)
	cloudPath := filepath.Base(file.Path) + "-" + strconv.Itoa(chunk.SequenceNumber)
	_, err = n.client.SendMessage(SaveChunkMsg, SaveChunkRequest{
		CloudPath: 	cloudPath,
		Chunk: 		chunk,
		Contents: 	contents,
	})
	return err
}

func (r request) OnSaveChunkRequest(sr SaveChunkRequest) error {
	utils.GetLogger().Printf("[INFO] Node: %v, received SaveChunk request.", r.cloud.MyNode.Name)
	// extract contents
	cloudPath := sr.CloudPath
	chunk := sr.Chunk
	contents := sr.Contents
	utils.GetLogger().Printf("[DEBUG] Got SaveChunkRequest with path: %v, chunk: %v.", cloudPath, chunk)
	err := r.cloud.saveChunk(cloudPath, chunk, contents)
	if err != nil {
		return err
	}
	err = r.node.updateChunkNodes(chunk.ID, r.cloud.MyNode.ID)
	return err
}

// Private requests and handlers.

// TODO: instead of sending entire ChunkNodes, only send the operation to be performed and a data item,
// i.e. addToChunkNodes(chunkID, nodeID)
// additionlly last call overrides all things
func (n *Node) updateChunkNodes(chunkID datastore.ChunkID, nodeID string) error {
	utils.GetLogger().Printf("[INFO] Sending updateChunkNodes request to node: %v.", n.Name)
	utils.GetLogger().Printf("[DEBUG] Sending message to client: %v.", &n.client)
	_, err := n.client.SendMessage(updateChunkNodesMsg, chunkID, nodeID)
	if err != nil {
		// FIXME: a way to propagate errors returned from requests, i.e. take the place of communication.go errors
		utils.GetLogger().Printf("[ERROR] %v.", err)
	}
	return err
}

func (r request) onUpdateChunkNodes(chunkID datastore.ChunkID, nodeID string) {
	utils.GetLogger().Printf("[INFO] Node: %v, received onUpdateChunkNodes request.", r.node.Name)
	utils.GetLogger().Printf("[DEBUG] Received request at client: %v.", &r.node.client)
	r.cloud.updateChunkNodes(chunkID, nodeID)
}

func createDataStoreRequestHandler(node *Node, cloud *Cloud) func(string) interface{} {
	utils.GetLogger().Printf("[INFO] Creating a datastore request handler for node: %v, and cloud: %v.", node, cloud)
	r := request{
		cloud: cloud,
		node: node,
	}
	utils.GetLogger().Printf("[DEBUG] Initialised request with fields: %v.", r)

	return func(message string) interface{} {
		switch message {
		case AddFileMsg: return r.OnAddFileRequest
		case SaveChunkMsg: return r.OnSaveChunkRequest
	    case updateChunkNodesMsg: return r.onUpdateChunkNodes
		}
		return nil
	}
}
