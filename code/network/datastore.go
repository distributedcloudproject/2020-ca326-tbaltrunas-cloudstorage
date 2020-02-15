package network

import (
	"cloud/datastore"
	"encoding/gob"
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
	r.cloud.Network.DataStore = append(r.cloud.Network.DataStore, file)
}

func (n *Node) SaveChunk(chunkID datastore.ChunkID, contents []byte) error {
	_, err := n.client.SendMessage(SaveChunkMsg, chunkID, contents)
	return err
}

func (r request) OnSaveChunkRequest(chunkID datastore.ChunkID, contents []byte) {
	// verify chunk ID

	// persistently store chunk

	// update chunk data structure
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
