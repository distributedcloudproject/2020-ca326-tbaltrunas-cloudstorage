package network

import (
	"cloud/utils"
	"encoding/gob"
)

const (
	NetworkInfoMsg = "NetworkInfo"
	NodeInfoMsg = "NodeInfo"
	AddNodeMsg = "AddNode"
)

func init() {
	gob.Register(Network{})
	gob.Register(Node{})
}

func createRequestHandler(node *Node, cloud *Cloud) func(string) interface{} {
	utils.GetLogger().Printf("[INFO] Creating a request handler for node: %v, cloud: %v.", node, cloud)
	r := request{
		cloud: cloud,
		node: node,
	}
	utils.GetLogger().Printf("[DEBUG] Initialied request with fields: %v.", r)

	return func(message string) interface{} {
		utils.GetLogger().Printf("[DEBUG] Anonymous handler switcher called for message: %v.", message)
		switch message {
		case "ping": return r.PingRequest
		case NetworkInfoMsg: return r.OnNetworkInfoRequest
		case NodeInfoMsg: return r.OnNodeInfoRequest
		case AddNodeMsg: return r.OnAddNodeRequest
		}
		return nil
	}
}

func (n *Node) NetworkInfo() (Network, error) {
	utils.GetLogger().Println("[INFO] Sending NetworkInfo request.")
	ret, err := n.client.SendMessage(NetworkInfoMsg)
	return ret[0].(Network), err
}

func (r request) OnNetworkInfoRequest() Network {
	utils.GetLogger().Println("[INFO] Handling NetworkInfo request.")
	r.cloud.Mutex.RLock()
	defer r.cloud.Mutex.RUnlock()
	return r.cloud.Network
}

func (n *Node) NodeInfo() (Node, error) {
	utils.GetLogger().Println("[INFO] Sending NodeInfo request.")
	ret, err := n.client.SendMessage(NodeInfoMsg)
	return ret[0].(Node), err
}

func (r request) OnNodeInfoRequest() Node {
	utils.GetLogger().Println("[INFO] Handling NodeInfo request.")
	r.node.mutex.RLock()
	defer r.node.mutex.RUnlock()
	return *r.node
}

func (n *Node) AddNode(node Node) error {
	utils.GetLogger().Printf("[INFO] Sending AddNode request with parameter node: %v.", node)
	_, err := n.client.SendMessage(AddNodeMsg, node)
	return err
}

func (r request) OnAddNodeRequest(node Node) {
	utils.GetLogger().Printf("[INFO] Handling AddNodeRequest with parameter node: %v.", node)
	r.cloud.addNode(&node)
	utils.GetLogger().Printf("[DEBUG] Added node to list of nodes: %v.", r.cloud.Network.Nodes)
}