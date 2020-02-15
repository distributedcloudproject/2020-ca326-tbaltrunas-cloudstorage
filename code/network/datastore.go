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

// func (n *Node) SaveChunk(chunk *datastore.Chunk, contents []byte, node *Node) error {
// 	_, err := SendMessage(SaveChunkMsg, chunk, contents, node)
// 	return err
// }

// func (r request) OnSaveChunkRequest(chunk *datastore.Chunk, contents []byte) {

// }

// LoadChunk(chunkID, node)

func createDataStoreRequestHandler(node *Node, cloud *Cloud) func(string) interface{} {
	r := request{
		cloud: cloud,
		node: node,
	}

	return func(message string) interface{} {
		switch message {
		case AddFileMsg: return r.OnAddFileRequest
		// case SaveChunkMsg: return r.OnSaveChunkRequest
		}
		return nil
	}
}
