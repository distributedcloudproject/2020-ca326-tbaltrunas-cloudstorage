package network

import (
	"cloud/utils"
	"cloud/datastore"
	"encoding/gob"
)
 	
const (
	AddFileMsg = "AddFile"
	SaveChunkMsg = "SaveChunk"
	updateFileChunkLocationsMsg = "updateFileChunkLocations"
)

type SaveChunkRequest struct {
	Path 		string			  // filepath corresponding to the chunk
	Chunk 		datastore.Chunk  // chunk metadata
	Contents 	[]byte           // chunk bytes
}

func init() {
	gob.Register(&datastore.File{})
	gob.Register(SaveChunkRequest{})
	gob.Register(FileChunkLocations{})
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
							 file, chunkNum, n)
	chunk, _, err := file.GetChunk(chunkNum)
	_, err = n.client.SendMessage(SaveChunkMsg, SaveChunkRequest{
		Path: 		file.Path,
		Chunk: 		file.Chunks.Chunks[chunkNum],
		Contents: 	chunk,
	})
	return err
}

func (r request) OnSaveChunkRequest(sr SaveChunkRequest) error {
	utils.GetLogger().Printf("[INFO] Node: %v, received SaveChunk request.", r.cloud.MyNode.ID)
	// extract contents
	path := sr.Path
	chunk := sr.Chunk
	contents := sr.Contents
	utils.GetLogger().Printf("[DEBUG] Got SaveChunkRequest with path: %v, chunk: %v.", path, chunk)
	err := r.cloud.saveChunk(path, chunk, contents)
	return err
}

// Private requests and handlers.

// TODO: instead of sending entire FileChunkLocations, only send the operation to be performed and a data item,
// i.e. addToFileChunkLocations(chunkID, nodeID)
func (n *Node) updateFileChunkLocations(fileChunkLocations FileChunkLocations) error {
	utils.GetLogger().Printf("[INFO] Sending updateFileChunkLocations request for FileChunkLocations: %v, on node: %v.", 
							 fileChunkLocations, n)
	_, err := n.client.SendMessage(updateFileChunkLocationsMsg, fileChunkLocations)
	if err != nil {
		// FIXME: a way to propagate errors returned from requests, i.e. take the place of communication.go errors
		utils.GetLogger().Printf("[ERROR] %v.", err)
	}
	return err
}

func (r request) onUpdateFileChunkLocations(fileChunkLocations FileChunkLocations) {
	utils.GetLogger().Printf("[INFO] Node: %v, received onUpdateFileChunkLocations request for FileChunkLocations: %v.", 
							 r.cloud.MyNode.ID, fileChunkLocations)
	r.cloud.updateFileChunkLocations(fileChunkLocations)
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
	    case updateFileChunkLocationsMsg: return r.onUpdateFileChunkLocations
		}
		return nil
	}
}

// Contains returns whether the datastore contains the specified file.
func (ds *DataStore) Contains(file *datastore.File) bool {
	for _, f := range ds.Files {
		// TODO: file ID
		if file.Path == f.Path {
			return true
		}
	}
	return false
}

// Add appends a file to the datastore.
func (ds *DataStore) Add(file *datastore.File) {
	ds.Files = append(ds.Files, file)
}

// TODO: move DataStore to datastore package
// Also DataStore is just a wrapper for a slice?
