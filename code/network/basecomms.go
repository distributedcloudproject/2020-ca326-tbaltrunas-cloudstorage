package network

import (
	"encoding/gob"
)

const (
	NetworkInfoMsg = "NetworkInfo"
	AddNodeMsg = "AddNode"
)

func init() {
	gob.Register(&Network{})
	gob.Register(&Node{})
}

func (n *Node) NetworkInfo() (*Network, error) {
	ret, err := n.client.SendMessage(NetworkInfoMsg)
	return ret[0].(*Network), err
}

func (r request) OnNetworkInfoRequest() *Network {
	r.network.mutex.RLock()
	defer r.network.mutex.RUnlock()
	return r.network
}

func (n *Node) AddNode(node Node) error {
	_, err := n.client.SendMessage(AddNodeMsg, node)
	return err
}

func (r request) OnAddNodeRequest(node Node) {
	r.network.mutex.Lock()
	defer r.network.mutex.Unlock()
	r.network.Nodes = append(r.network.Nodes, &node)
}