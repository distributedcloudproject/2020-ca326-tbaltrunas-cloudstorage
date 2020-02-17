package network

import (
	"cloud/utils"
	"cloud/datastore"
	"encoding/gob"
	"os"
	"strconv"
	"path/filepath"
)
 	
const (
	AddFileMsg = "AddFile"
	SaveChunkMsg = "SaveChunk"
)

type SaveChunkRequest struct {
	Path 		string			  // filepath corresponding to the chunk
	Chunk 		datastore.Chunk  // chunk metadata
	Contents 	[]byte           // chunk bytes
}

func init() {
	gob.Register(&datastore.File{})
	gob.Register(SaveChunkRequest{})
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
	if err != nil {
		return err 
	}
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

	// verify chunk ID
	chunkID := chunk.ID

	// persistently store chunk
	// r.cloud.Mutex.Lock()
	// defer r.cloud.Mutex.Unlock()
	
	chunkPath := filepath.Join(r.cloud.MyNode.FileStorageDir, filepath.Dir(path), 
								filepath.Base(path) + "-" + strconv.Itoa(chunk.SequenceNumber))
	utils.GetLogger().Printf("[DEBUG] Computed path where to store chunk: %s.", chunkPath)
	// TODO: maybe this should be done when setting up the node
	err := os.MkdirAll(filepath.Dir(chunkPath), os.ModeDir)
	if err != nil {
		return err
	}
	utils.GetLogger().Println("[DEBUG] Created/verified existence of path directories.")
	w, err := os.Create(chunkPath)
	if err != nil {
		return err 
	}
	defer w.Close()

	utils.GetLogger().Printf("[DEBUG] Saving chunk to writer: %v.", w)
	err = datastore.SaveChunk(w, contents)
	if err != nil { 
		return err
	}
	utils.GetLogger().Printf("[DEBUG] Finished saving chunk to writer: %v.", w)

	// update chunk data structure
	utils.GetLogger().Println("[DEBUG] Updating FileChunkLocations.")
	nodeID := r.cloud.MyNode.ID
	chunkLocations, ok := r.cloud.Network.FileChunkLocations[chunkID]
	if !ok {
		chunkLocations = []string{nodeID}
	} else {
		chunkLocations = append(chunkLocations, nodeID)
	}
	r.cloud.Network.FileChunkLocations[chunkID] = chunkLocations
	utils.GetLogger().Println("[DEBUG] Finished updating FileChunkLocations.")
	return nil
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
		}
		return nil
	}
}
