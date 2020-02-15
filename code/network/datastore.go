package network

import (
	"cloud/datastore"
	"encoding/gob"
	"os"
)

const (
	AddFileMsg = "AddFile"
	SaveChunkMsg = "SaveChunk"
)

func init() {
	gob.Register(datastore.File{})
}

func (n *Node) AddFile(file *datastore.File) error {
	_, err := n.client.SendMessage(AddFileMsg, file)
	return err
}

func (r request) OnAddFileRequest(file *datastore.File) {
	r.cloud.Network.DataStore.Files = append(r.cloud.Network.DataStore.Files, file)
}

func (n *Node) SaveChunk(chunkID datastore.ChunkID, contents []byte) error {
	_, err := n.client.SendMessage(SaveChunkMsg, chunkID, contents)
	return err
}

func (r request) OnSaveChunkRequest(chunkID datastore.ChunkID, contents []byte) error {
	// verify chunk ID

	// get file belonging to chunk
	// TODO: pass file, or the path where the chunk should be stored?
	file := r.cloud.Network.DataStore.GetFileByChunkID(chunkID)

	// persistently store chunk
	r.cloud.Mutex.RLock()
	defer r.cloud.Mutex.RUnlock()
	// TODO: decide path of the chunk
	w, err := os.Create("/tmp/cloud_chunk_saved")
	if err != nil { return err }
	err = file.SaveChunk(w, contents)
	if err != nil { return err }

	// TODO: mutex

	// update chunk data structure
	nodeID := r.cloud.MyNode.ID
	nodesWithChunk, ok := r.cloud.Network.FileChunkLocations[chunkID]
	if !ok {
		nodesWithChunk = []string{nodeID}
	} else {
		nodesWithChunk = append(nodesWithChunk, nodeID)
	}
	r.cloud.Network.FileChunkLocations[chunkID] = nodesWithChunk
	return nil
}

func createDataStoreRequestHandler(node *Node, cloud *Cloud) func(string) interface{} {
	r := request{
		cloud: cloud,
		node: node,
	}

	return func(message string) interface{} {
		switch message {
		case AddFileMsg: return r.OnAddFileRequest
		case SaveChunkMsg: return r.OnSaveChunkRequest
		}
		return nil
	}
}

func (ds *DataStore) GetFileByChunkID(chunkID datastore.ChunkID) *datastore.File {
	for _, file := range ds.Files {
		for _, chunk := range file.Chunks.Chunks {
			if chunk.ID == chunkID {
				return file
			}
		}
	}
	return nil
}
