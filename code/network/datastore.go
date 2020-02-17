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
	utils.GetLogger().Printf("[DEBUG] client is: %v.", n.client)
	_, err := n.client.SendMessage(AddFileMsg, file)
	return err
}

func (r request) OnAddFileRequest(file *datastore.File) {
	utils.GetLogger().Printf("[INFO] Received AddFile request for file: %v.", file)
	r.cloud.Mutex.Lock()
	defer r.cloud.Mutex.Unlock()
	r.cloud.Network.DataStore.Files = append(r.cloud.Network.DataStore.Files, file)
	r.cloud.Save()
}

func (n *Node) SaveChunk(file *datastore.File, chunkNum int) error {
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
	// extract contents
	path := sr.Path
	chunk := sr.Chunk
	contents := sr.Contents

	// verify chunk ID
	chunkID := chunk.ID

	// persistently store chunk
	r.cloud.Mutex.RLock()
	defer r.cloud.Mutex.RUnlock()
	
	// TODO: decide path of the chunk
	chunkPath := filepath.Join(r.cloud.MyNode.FileStorageDir, filepath.Dir(path), 
								filepath.Base(path) + "-" + strconv.Itoa(chunk.SequenceNumber))
	err := os.MkdirAll(filepath.Dir(chunkPath), os.ModeDir)
	if err != nil {
		return err
	}
	w, err := os.Create(chunkPath)
	if err != nil {
		return err 
	}
	defer w.Close()

	err = datastore.SaveChunk(w, contents)
	if err != nil { 
		return err
	}

	// update chunk data structure
	nodeID := r.cloud.MyNode.ID
	chunkLocations, ok := r.cloud.Network.FileChunkLocations[chunkID]
	if !ok {
		chunkLocations = []string{nodeID}
	} else {
		chunkLocations = append(chunkLocations, nodeID)
	}
	r.cloud.Network.FileChunkLocations[chunkID] = chunkLocations
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
