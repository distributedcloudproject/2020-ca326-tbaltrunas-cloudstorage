package network

import (
	"cloud/datastore"
	"encoding/gob"
	"fmt"
)

const (
	AddFileMsg = "AddFile"
)

func init() {
	gob.Register(datastore.File{})
}

func (n *Node) AddFile(file *datastore.File) error {
	fmt.Println(file)
	fmt.Println("start")
	_, err := n.client.SendMessage(AddFileMsg, file)
	fmt.Println("finish")
	return err
}

func (r request) OnAddFileRequest(file *datastore.File) {
	fmt.Println(file)
	// c.Network.DataStore = append(c.Network.DataStore, file)
}

// SaveChunk(chunk, node)

// LoadChunk(chunkID, node)

func createDataStoreRequestHandler(node *Node, cloud *Cloud) func(string) interface{} {
	r := request{
		cloud: cloud,
		node: node,
	}

	return func(message string) interface{} {
		switch message {
		case AddFileMsg: return r.OnAddFileRequest
		}
		return nil
	}
}
