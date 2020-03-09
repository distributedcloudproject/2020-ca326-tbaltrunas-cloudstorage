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
	"path/filepath"
	"strings"
	"path/filepath"
)

type ChunkNodes map[datastore.ChunkID][]string
type FileNodes map[datastore.FileID][]string

type request struct {
	Cloud    *cloud
	FromNode *cloudNode
}

// NetworkFile is used for returning a File variable together with its full path.
type NetworkFile struct {
	File *datastore.File
	Path string
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

// CleanNetworkPath cleans the provided path and returns a network-friendly path. Always starting with a / and only
// containing forward slashes.
func CleanNetworkPath(networkPath string) string {
	networkPath = filepath.ToSlash(networkPath)
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

func (c *cloud) GetFolders() []*NetworkFolder {
	c.networkMutex.RLock()
	defer c.networkMutex.RUnlock()
	return c.network.GetFolders()
}

func (c *cloud) GetFiles() []*NetworkFile {
	c.networkMutex.RLock()
	defer c.networkMutex.RUnlock()
	return c.network.GetFiles()
}

// GetFile retrieves the metadata of a file given it's path.
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

// GetFolder retrieves the folder for the given path.
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

// GetFolders retrieve the folders in the network.
func (n *Network) GetFolders() []*NetworkFolder {
	return n.rGetFolders(n.RootFolder)
}

func (n *Network) rGetFolders(folder *NetworkFolder) []*NetworkFolder {
	folders := make([]*NetworkFolder, 0)

	// base case
	if folder == nil {
		return folders
	}

	// recursive case
	folders = append(folders, folder)
	for _, subfolder := range folder.SubFolders {
		subfolders := n.rGetFolders(subfolder)
		folders = append(folders, subfolders...)
	}
	return folders
}

// GetFiles retrieve the files in the network.
func (n *Network) GetFiles() []*datastore.File {
	return n.rGetFiles(n.RootFolder)
}

func (n *Network) rGetFiles(folder *NetworkFolder) []*NetworkFile {
	netFiles := make([]*NetworkFile, 0)
	// base case
	if folder == nil {
		return netFiles
	}

	// recursive case
	for _, file := range folder.Files.Files {
		netFiles = append(netFiles, &NetworkFile{
			File: file,
			Path: filepath.Join(folder.Name, file.Name),
		})
	}
	for _, subfolder := range folder.SubFolders {
		netSubfiles := n.rGetFiles(subfolder)
		netFiles = append(netFiles, netSubfiles...)
	}
	return netFiles
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

// PublicKeyToID generates a sha256 hash based on the public key.
func PublicKeyToID(key *rsa.PublicKey) (string, error) {
	pub, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return "", err
	}
	sha := sha256.Sum256(pub)
	return hex.EncodeToString(sha[:]), nil
}
