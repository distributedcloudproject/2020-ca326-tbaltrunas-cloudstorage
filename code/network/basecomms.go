package network

import (
	"cloud/datastore"
	"encoding/gob"
	"fmt"
)

const (
	NetworkInfoMsg = "NetworkInfo"
	NodeInfoMsg = "NodeInfo"
	AddNodeMsg = "AddNode"
	AddFileMsg = "AddFile"
)

func init() {
	gob.Register(Network{})
	gob.Register(Node{})
	gob.Register(datastore.File{})
}

func createRequestHandler(node *Node, cloud *Cloud) func(string) interface{} {
	r := request{
		cloud: cloud,
		node: node,
	}

	return func(message string) interface{} {
		switch message {
		case "ping": return r.PingRequest
		case NetworkInfoMsg: return r.OnNetworkInfoRequest
		case NodeInfoMsg: return r.OnNodeInfoRequest
		case AddNodeMsg: return r.OnAddNodeRequest
		case AddFileMsg: return r.OnAddFileRequest
		}
		return nil
	}
}

func (n *Node) NetworkInfo() (Network, error) {
	ret, err := n.client.SendMessage(NetworkInfoMsg)
	return ret[0].(Network), err
}

func (r request) OnNetworkInfoRequest() Network {
	r.cloud.Mutex.RLock()
	defer r.cloud.Mutex.RUnlock()
	return r.cloud.Network
}

func (n *Node) NodeInfo() (Node, error) {
	ret, err := n.client.SendMessage(NodeInfoMsg)
	return ret[0].(Node), err
}

func (r request) OnNodeInfoRequest() Node {
	r.node.mutex.RLock()
	defer r.node.mutex.RUnlock()
	return *r.node
}

func (n *Node) AddNode(node Node) error {
	_, err := n.client.SendMessage(AddNodeMsg, node)
	return err
}

func (r request) OnAddNodeRequest(node Node) {
	r.cloud.Mutex.Lock()
	defer r.cloud.Mutex.Unlock()

	for _, n := range r.cloud.Network.Nodes {
		if n.ID == node.ID {
			if n.client == nil {
				n.IP = node.IP
				n.Name = node.Name
				n.client = node.client
			}
			return
		}
	}
	r.cloud.Network.Nodes = append(r.cloud.Network.Nodes, &node)
	r.cloud.Save()
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
