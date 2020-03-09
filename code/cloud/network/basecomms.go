package network

import (
	"cloud/utils"
	"crypto/rsa"
	"encoding/gob"
	"errors"
)

// Messages used for basic communication.
const (
	NetworkInfoMsg    = "NetworkInfo"
	NodeInfoMsg       = "NodeInfo"
	AddNodeMsg        = "AddNode"
	AddToWhitelist    = "AddToWhitelist"
	RemoveToWhitelist = "RemoveFromWhitelist"
)

func init() {
	gob.Register(Network{})
	gob.Register(Node{})
	gob.Register(rsa.PublicKey{})

	handlers = append(handlers, createRequestHandler)
}

func createRequestHandler(node *cloudNode, cloud *cloud) func(string) interface{} {
	utils.GetLogger().Printf("[INFO] Creating a request handler for node: %v, cloud: %v.", node, cloud)
	r := request{
		Cloud:    cloud,
		FromNode: node,
	}
	utils.GetLogger().Printf("[DEBUG] Initialied request with fields: %v.", r)

	return func(message string) interface{} {
		utils.GetLogger().Printf("[DEBUG] Anonymous handler switcher called for message: %v.", message)
		switch message {
		case "ping":
			return r.PingRequest
		case NetworkInfoMsg:
			return r.OnNetworkInfoRequest
		case NodeInfoMsg:
			return r.OnNodeInfoRequest
		case AddNodeMsg:
			return r.OnAddNodeRequest
		case AddToWhitelist:
			return r.OnAddToWhitelist
		case RemoveToWhitelist:
			return r.OnRemoveFromWhitelist
		}
		return nil
	}
}

func (n *cloudNode) NetworkInfo() (Network, error) {
	utils.GetLogger().Println("[INFO] Sending NetworkInfo request.")
	ret, err := n.client.SendMessage(NetworkInfoMsg)
	return ret[0].(Network), err
}

func (r request) OnNetworkInfoRequest() Network {
	utils.GetLogger().Println("[INFO] Handling NetworkInfo request.")
	r.Cloud.networkMutex.RLock()
	defer r.Cloud.networkMutex.RUnlock()
	return r.Cloud.network
}

func (n *cloudNode) NodeInfo() (Node, error) {
	utils.GetLogger().Println("[INFO] Sending NodeInfo request.")
	ret, err := n.client.SendMessage(NodeInfoMsg)
	return ret[0].(Node), err
}

func (r request) OnNodeInfoRequest() (Node, error) {
	utils.GetLogger().Println("[INFO] Handling NodeInfo request.")
	n, ok := r.Cloud.NodeByID(r.Cloud.MyNode().ID)
	if !ok {
		return Node{}, errors.New("could not find node")
	}
	return n, nil
}

func (c *cloud) AddNode(node Node) {
	c.SendMessageToMe(AddNodeMsg, node)
	c.SendMessageAllOthers(AddNodeMsg, node)
}

func (r request) OnAddNodeRequest(node Node) {
	utils.GetLogger().Printf("[INFO] Handling AddNodeRequest with parameter node: %v.", node)

	r.Cloud.networkMutex.Lock()
	defer r.Cloud.networkMutex.Unlock()

	for i := range r.Cloud.network.Nodes {
		// If there is a matching node, update instead.
		if r.Cloud.network.Nodes[i].ID == node.ID {
			r.Cloud.network.Nodes[i] = node

			if r.Cloud.events.NodeUpdated != nil {
				go r.Cloud.events.NodeUpdated(node)
			}
			return
		}
	}
	r.Cloud.network.Nodes = append(r.Cloud.network.Nodes, node)

	if r.Cloud.events.NodeAdded != nil || r.Cloud.events.NodeConnected != nil {
		go func() {
			if r.Cloud.events.NodeAdded != nil {
				r.Cloud.events.NodeAdded(node)
			}

			if r.Cloud.IsNodeOnline(node.ID) && r.Cloud.events.NodeConnected != nil {
				r.Cloud.events.NodeConnected(node.ID)
			}
		}()
	}
}

func (c *cloud) AddToWhitelist(ID string) error {
	_, err := c.SendMessageToMe(AddToWhitelist, ID)
	if err != nil {
		return err
	}
	go c.SendMessageAllOthers(AddToWhitelist, ID)
	return nil
}

func (r request) OnAddToWhitelist(ID string) error {
	utils.GetLogger().Printf("[DEBUG] Added ID to list of nodes: %v.", ID)
	if ID == "" {
		return errors.New("cannot add empty ID")
	}

	r.Cloud.networkMutex.Lock()
	defer r.Cloud.networkMutex.Unlock()

	for i := range r.Cloud.network.WhitelistIDs {
		if r.Cloud.network.WhitelistIDs[i] == ID {
			return errors.New("ID is already whitelisted")
		}
	}
	r.Cloud.network.WhitelistIDs = append(r.Cloud.network.WhitelistIDs, ID)

	if r.Cloud.events.WhitelistAdded != nil {
		go r.Cloud.events.WhitelistAdded(ID)
	}

	return nil
}

func (c *cloud) RemoveFromWhitelist(ID string) error {
	_, err := c.SendMessageToMe(RemoveToWhitelist, ID)
	if err != nil {
		return err
	}
	go c.SendMessageAllOthers(RemoveToWhitelist, ID)
	return nil
}

func (r request) OnRemoveFromWhitelist(ID string) error {
	if ID == "" {
		return errors.New("cannot add empty ID")
	}

	r.Cloud.networkMutex.Lock()
	defer r.Cloud.networkMutex.Unlock()

	for i := range r.Cloud.network.WhitelistIDs {
		if r.Cloud.network.WhitelistIDs[i] == ID {
			r.Cloud.network.WhitelistIDs = append(r.Cloud.network.WhitelistIDs[:i], r.Cloud.network.WhitelistIDs[i+1:]...)
			if r.Cloud.events.WhitelistRemoved != nil {
				go r.Cloud.events.WhitelistRemoved(ID)
			}
			return nil
		}
	}
	return errors.New("ID is not whitelisted")
}
