package network

import (
	"cloud/comm"
	"cloud/utils"
	"crypto"
)

// Node is a global representation of any node. Each network will have the same view of the node.
type Node struct {
	// Unique ID of the Node.
	ID string

	// Represented in ip:port format.
	// Example: 127.0.0.1:8081
	IP string

	// Display name of the node.
	Name string

	// Public key of the node.
	PublicKey crypto.PublicKey
}

// cloudNode is the client's view of any node. This is unique to each node.
type cloudNode struct {
	ID     string
	client comm.Client
}

func (c *cloud) IsNodeOnline(ID string) bool {
	utils.GetLogger().Println("[DEBUG] Checking if node is online.")
	return c.hasCloudNode(ID)
}

func (c *cloud) addCloudNode(ID string, node *cloudNode) bool {
	c.NodesMutex.Lock()
	defer c.NodesMutex.Unlock()
	if _, ok := c.Nodes[ID]; !ok {
		c.Nodes[ID] = node

		if c.events.NodeConnected != nil {
			go c.events.NodeConnected(ID)
		}
		return true
	}
	return false
}

func (c *cloud) removeCloudNode(ID string) bool {
	c.NodesMutex.Lock()
	defer c.NodesMutex.Unlock()
	if _, ok := c.Nodes[ID]; ok {
		delete(c.Nodes, ID)

		if c.events.NodeDisconnected != nil {
			go c.events.NodeDisconnected(ID)
		}
		return true
	}
	return false
}

func (c *cloud) hasCloudNode(ID string) bool {
	c.NodesMutex.RLock()
	defer c.NodesMutex.RUnlock()
	_, ok := c.Nodes[ID]
	return ok
}

func (c *cloud) GetCloudNode(ID string) *cloudNode {
	c.NodesMutex.RLock()
	defer c.NodesMutex.RUnlock()
	return c.Nodes[ID]
}

func (c *cloud) removePendingNode(client comm.Client) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	for i := 0; i < len(c.PendingNodes); i++ {
		if c.PendingNodes[i].client == client {
			c.PendingNodes[i] = c.PendingNodes[len(c.PendingNodes)-1]
			c.PendingNodes = c.PendingNodes[:len(c.PendingNodes)]
		}
	}
}
