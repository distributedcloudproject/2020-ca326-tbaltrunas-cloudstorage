package network

import (
	"cloud/datastore"
	"cloud/utils"
	"crypto/rsa"
	"encoding/gob"
	"os"
	"path/filepath"
)

func init() {
	gob.Register(SavedNetworkState{})
}

// SavedNetworkState contains information to re-spawn a network from an offline state.
type SavedNetworkState struct {
	Network Network
	Config  CloudConfig

	MyNode     Node
	PrivateKey *rsa.PrivateKey

	FileStorage map[string]datastore.FileStore
	FileSyncs   []*datastore.SyncFileStore
	FolderSyncs []fileSync
}

func (c *cloud) SavedNetworkState() SavedNetworkState {
	utils.GetLogger().Println("[INFO] Retrieving Saved Network State.")
	c.networkMutex.RLock()
	defer c.networkMutex.RUnlock()
	c.Mutex.RLock()
	defer c.Mutex.RUnlock()
	//fmt.Println(c.folderSyncs)
	return SavedNetworkState{
		Network:     c.network,
		Config:      c.Config(),
		MyNode:      c.myNode,
		PrivateKey:  c.privateKey,
		FileStorage: c.fileStorage,
		FileSyncs:   c.fileSyncs,
		FolderSyncs: c.folderSyncs,
	}
}

// LoadNetwork uses a SavedNetworkState to connect to the network, if it's up. If the network is offline, it will bring
// it back online.
func LoadNetwork(s SavedNetworkState) Cloud {
	utils.GetLogger().Println("[INFO] Loading cloud network.")

	for _, n := range s.Network.Nodes {
		c, err := BootstrapToNetwork(n.IP, s.MyNode, s.PrivateKey, s.Config)
		if err != nil {
			continue
		}
		return c
	}
	utils.GetLogger().Println("[INFO] Could not reconnect to the network. Starting our own.")
	c := SetupNetwork(s.Network, s.MyNode, s.PrivateKey)
	c.SetConfig(s.Config)
	cc := c.(*cloud)
	if s.FileStorage != nil {
		cc.fileStorage = s.FileStorage
	}
	cc.fileSyncs = s.FileSyncs
	cc.folderSyncs = s.FolderSyncs
	//fmt.Println(cc.folderSyncs, s.FolderSyncs)
	cc.createWatcher()
	for _, f := range cc.fileSyncs {
		cc.watcher.Add(f.FilePath)
	}
	for _, f := range cc.folderSyncs {
		cc.watcher.Add(f.LocalPath)
		filepath.Walk(f.LocalPath, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				cc.watcher.Add(path)
			}
			return nil
		})
	}
	return c
}
