package network

import (
	"encoding/gob"
)

const (
	NetworkInfoMsg = "NetworkInfo"
	NodeInfoMsg = "NodeInfo"
	AddNodeMsg = "AddNode"
	AddToWhitelist = "AddToWhitelist"
)

func init() {
	gob.Register(Network{})
	gob.Register(Node{})
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
		case AddToWhitelist: return r.OnAddToWhitelist
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
	r.cloud.addNode(&node)
}

func (n *Node) AddToWhitelist(ID string) error {
	_, err := n.client.SendMessage(AddToWhitelist, ID)
	return err
}

func (r request) OnAddToWhitelist(ID string) {
	r.cloud.addToWhitelist(ID)
}