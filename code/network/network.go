package network

import (
	"cloud/datastore"
	"cloud/utils"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
)

// Network is the general info of the network. Each node would have the same presentation of Network.
type Network struct {
	Name  string
	Nodes []Node

	// Require authentication for the network. Authentication verifies that Node ID belongs to the public key.
	RequireAuth bool

	// Enable whitelist for the network. If enabled, Node ID has to be whitelisted before joining the network.
	Whitelist bool

	// List of node IDs that are permitted to enter the network.
	WhitelistIDs []string

	// DataStore is a list of all the user files on the cloud.
	DataStore datastore.DataStore

	// ChunkNodes maps chunk ID's to the Nodes (Node ID's) that contain that chunk.
	// This way we can keep track of which nodes contain which chunks.
	// And make decisions about the chunk requets to perform.
	// In the future this scheme might change, for example, with each node knowing only about its own chunks.
	ChunkNodes ChunkNodes
}

type ChunkNodes map[datastore.ChunkID][]string

type request struct {
	Cloud    *cloud
	FromNode *cloudNode
}

func (c *cloud) NodeByID(ID string) (node Node, found bool) {
	c.networkMutex.RLock()
	defer c.networkMutex.RUnlock()
	return c.network.NodeByID(ID)
}

func (n *Network) NodeByID(ID string) (node Node, found bool) {
	for i := 0; i < len(n.Nodes); i++ {
		if n.Nodes[i].ID == ID {
			return n.Nodes[i], true
		}
	}
	return Node{}, false
}

func (c *cloud) IsWhitelisted(ID string) bool {
	if ID == "" {
		return false
	}

	c.networkMutex.RLock()
	defer c.networkMutex.RUnlock()

	for i := range c.network.WhitelistIDs {
		if c.network.WhitelistIDs[i] == ID {
			return true
		}
	}
	return false
}

func (c *cloud) Whitelist() []string {
	c.networkMutex.RLock()
	defer c.networkMutex.RUnlock()

	wl := make([]string, len(c.network.WhitelistIDs))
	copy(wl, c.network.WhitelistIDs)
	return wl
}

func (n *cloudNode) Ping() (string, error) {
	utils.GetLogger().Println("[INFO] Pinging node.")
	ping, err := n.client.SendMessage("ping", "ping")
	if err != nil {
		return "", err
	}
	return ping[0].(string), err
}

func (r request) PingRequest(ping string) string {
	utils.GetLogger().Printf("[INFO] Handling ping request with string: %v.", ping)
	if ping == "ping" {
		return "pong"
	}
	return ""
}

func PublicKeyToID(key *rsa.PublicKey) (string, error) {
	pub, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return "", err
	}
	sha := sha256.Sum256(pub)
	return hex.EncodeToString(sha[:]), nil
}
