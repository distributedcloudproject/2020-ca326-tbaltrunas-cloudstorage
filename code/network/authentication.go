package network

import (
	"cloud/utils"
	"encoding/gob"
	"fmt"
	"strings"
)

const (
	AuthMsg = "authenticate_me"
)

type AuthRequest struct {
	ID   string
	IP   string
	Name string
}

func init() {
	gob.Register(AuthRequest{})
}

func (n *cloudNode) Authenticate(node Node) (bool, error) {
	utils.GetLogger().Printf("[INFO] Sending Authenticate request with parameter node: %v.", node)
	success, err := n.client.SendMessage(AuthMsg, AuthRequest{
		ID:   node.ID,
		IP:   node.IP,
		Name: node.Name,
	})
	return success[0].(bool), err
}

func (r request) OnAuthenticateRequest(ar AuthRequest) bool {
	utils.GetLogger().Printf("[INFO] Handling Authenticate request with AuthRequest struct parameter: %v.", ar)

	// Format the IP correctly.
	if ar.IP == "" {
		ar.IP = r.FromNode.client.Address()
	} else if ar.IP[0] == ':' {
		ip := strings.Split(r.FromNode.client.Address(), ":")
		ar.IP = ip[0] + ar.IP
	}
	utils.GetLogger().Printf("[DEBUG] Got full IP in auth struct: %v.", ar.IP)

	// Create a Node for the request.
	node := Node{
		ID:        ar.ID,
		IP:        ar.IP,
		Name:      ar.Name,
		PublicKey: r.FromNode.client.PublicKey(),
	}
	utils.GetLogger().Printf("[DEBUG] Updated context request node: %v.", r)

	// Verify the ID belongs to the public key.
	id, err := PublicKeyToID(r.FromNode.client.PublicKey())
	if err != nil {
		return false
	}
	if id != node.ID {
		fmt.Println("BAD KEY", id, node.ID)
		return false
	}
	r.FromNode.ID = id

	// If whitelist is enabled, verify that the node is allowed to access it.
	if r.Cloud.network.Whitelist {
		if !r.Cloud.IsWhitelisted(id) {
			// LOG: fmt.Println("Unauthorized access", id)
			return false
		}
		go r.Cloud.RemoveFromWhitelist(id)
	}

	// Add any request handlers - node is now part of the network.
	r.Cloud.addRequestHandlers(r.FromNode)
	utils.GetLogger().Printf("[DEBUG] Added a request handler for node's client: %v.", r.FromNode.client)

	// Remove the node from the pending nodes list.
	r.Cloud.removePendingNode(r.FromNode.client)

	// Add the node to the network.
	r.Cloud.AddNode(node)

	// Add the node to our online nodes.
	r.Cloud.addCloudNode(node.ID, r.FromNode)

	utils.GetLogger().Println("[DEBUG] Adding new node and updating clients.")

	return true
}

func createAuthRequestHandler(cNode *cloudNode, cloud *cloud) func(string) interface{} {
	utils.GetLogger().Printf("[INFO] Creating an auth request handler for node: %v, and cloud: %v.", cNode.ID, cloud)
	r := request{
		Cloud:    cloud,
		FromNode: cNode,
	}
	utils.GetLogger().Printf("[DEBUG] Initialised request with fields: %v.", r)

	return func(message string) interface{} {
		switch message {
		case AuthMsg:
			return r.OnAuthenticateRequest
		}
		return nil
	}
}
