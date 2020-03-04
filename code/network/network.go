package network

import (
	"cloud/datastore"
	"cloud/utils"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"path"
	"strings"
)

type ChunkNodes map[datastore.ChunkID][]string
type FileNodes map[datastore.FileID][]string

type request struct {
	Cloud    *cloud
	FromNode *cloudNode
}

type NetworkFolder struct {
	Name       string
	SubFolders []*NetworkFolder

	// Files is a list of files in current folder on the Cloud.
	Files datastore.DataStore

	// ChunkNodes maps chunk ID's to the Nodes (Node ID's) that contain that chunk.
	// This way we can keep track of which nodes contain which chunks.
	// And make decisions about the chunk requests to perform.
	// In the future this scheme might change, for example, with each node knowing only about its own chunks.
	// ChunkNodes ChunkNodes
}

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

	RootFolder *NetworkFolder

	// ChunkNodes maps chunk ID's to the Nodes (Node ID's) that contain that chunk.
	// This way we can keep track of which nodes contain which chunks.
	// And make decisions about the chunk requests to perform.
	// In the future this scheme might change, for example, with each node knowing only about its own chunks.
	ChunkNodes ChunkNodes

	// FileNodes maps file ID's to the Nodes that contain the whole file. Those nodes are syncing the whole file all the
	// time.
	FileNodes FileNodes
}

func CleanNetworkPath(networkPath string) string {
	networkPath = path.Clean(networkPath)
	if len(networkPath) == 0 {
		return "/"
	}
	if len(networkPath) == 1 && networkPath[0] == '.' {
		return "/"
	}
	if len(networkPath) > 1 && networkPath[0] == '.' && networkPath[1] == '/' {
		return networkPath[1:]
	}
	if networkPath[0] != '/' {
		return "/" + networkPath
	}
	return networkPath
}

func (c *cloud) GetFolder(folder string) (*NetworkFolder, error) {
	c.networkMutex.RLock()
	defer c.networkMutex.RUnlock()
	return c.network.GetFolder(folder)
}

func (c *cloud) GetFile(file string) (*datastore.File, error) {
	c.networkMutex.RLock()
	defer c.networkMutex.RUnlock()
	return c.network.GetFile(file)
}

func (n *Network) GetFile(file string) (*datastore.File, error) {
	dir, base := path.Split(file)
	folder, err := n.GetFolder(dir)
	if err != nil {
		return nil, err
	}

	for _, f := range folder.Files.Files {
		if f.Name == base {
			return f, nil
		}
	}
	return nil, errors.New("file not found")
}

func (n *Network) GetFolder(folder string) (*NetworkFolder, error) {
	paths := strings.Split(folder, "/")

	if n.RootFolder == nil {
		n.RootFolder = &NetworkFolder{
			Name: "/",
		}
	}
	f := n.RootFolder

	for _, p := range paths {
		if p == "" {
			continue
		}

		// TODO: Check validity of path.

		foundFolder := false
		for _, sub := range f.SubFolders {
			if sub.Name == p {
				foundFolder = true
				f = sub
				break
			}
		}
		if !foundFolder {
			newFolder := &NetworkFolder{Name: p, Files: datastore.DataStore{}}
			f.SubFolders = append(f.SubFolders, newFolder)
			f = newFolder
		}
	}
	return f, nil
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
