package network

import (
	"cloud/utils"
	"encoding/gob"
	"strings"
)

const (
	AuthMsg = "authenticate_me"
)

type AuthRequest struct {
	ID string
	IP string
	Name string
}

func init() {
	gob.Register(AuthRequest{})
}

func (n *Node) Authenticate(node *Node) error {
	utils.GetLogger().Printf("[INFO] Sending Authenticate request with parameter node: %v.", node)
	_, err := n.client.SendMessage(AuthMsg, AuthRequest{
		ID: node.ID,
		IP: node.IP,
		Name: node.Name,
	})
	return err
}

func (r request) OnAuthenticateRequest(ar AuthRequest) {
	utils.GetLogger().Printf("[INFO] Handling Authenticate request with AuthRequest struct parameter: %v.", ar)
	r.node.mutex.Lock()
	defer r.node.mutex.Unlock()

	if ar.IP == "" {
		ar.IP = r.node.client.Address()
	}else if ar.IP[0] == ':' {
		ip := strings.Split(r.node.client.Address(), ":")
		ar.IP = ip[0] + ar.IP
	}
	utils.GetLogger().Printf("[DEBUG] Got full IP in auth struct: %v.", ar.IP)

	// Update our information on the node.
	r.node.ID = ar.ID
	r.node.IP = ar.IP
	r.node.Name = ar.Name
	utils.GetLogger().Printf("[DEBUG] Updated context request node: %v.", r)

	// Add any request handlers - node is now part of the network.
	r.node.client.AddRequestHandler(createRequestHandler(r.node, r.cloud))
	r.node.client.AddRequestHandler(createDataStoreRequestHandler(r.node, r.cloud))
	utils.GetLogger().Printf("[DEBUG] Added request handlers for node's client: %v.", r.node.client)

	// Remove the node from the pending nodes list.
	r.cloud.Mutex.Lock()
	utils.GetLogger().Printf("[DEBUG] Checking pending nodes for request node: %v.", r.cloud.PendingNodes)
	for i := 0; i < len(r.cloud.PendingNodes); i++ {
		if r.cloud.PendingNodes[i] == r.node {
			r.cloud.PendingNodes[i] = r.cloud.PendingNodes[len(r.cloud.PendingNodes) - 1]
			r.cloud.PendingNodes = r.cloud.PendingNodes[:len(r.cloud.PendingNodes)]
		}
	}
	utils.GetLogger().Printf("[DEBUG] Updated pending nodes for request node: %v.", r.cloud.PendingNodes)
	r.cloud.Mutex.Unlock()

	utils.GetLogger().Println("[DEBUG] Adding new node and updating clients.")
	// Tell all of the other nodes to add the node.
	for i := 0; i < len(r.cloud.Network.Nodes); i++ {
		if r.cloud.Network.Nodes[i].client != nil {
			go r.cloud.Network.Nodes[i].AddNode(*r.node)
		}
	}

	// Add the node to our list.
	r.cloud.addNode(r.node)
}

func createAuthRequestHandler(node *Node, cloud *Cloud) func(string) interface{} {
	utils.GetLogger().Printf("[INFO] Creating an auth request handler for node: %v, and cloud: %v.", node, cloud)
	r := request{
		cloud: cloud,
		node: node,
	}
	utils.GetLogger().Printf("[DEBUG] Initialised request with fields: %v.", r)

	return func(message string) interface{} {
		switch message {
		case AuthMsg: return r.OnAuthenticateRequest
		}
		return nil
	}
}