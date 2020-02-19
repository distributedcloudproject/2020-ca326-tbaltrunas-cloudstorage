package network

import (
	"cloud/utils"
	"crypto/rsa"
)

type SavedNetworkState struct {
	Network Network

	MyNode     Node
	PrivateKey *rsa.PrivateKey
}

func (c *cloud) SavedNetworkState() SavedNetworkState {
	utils.GetLogger().Println("[INFO] Retrieving Saved Network State.")
	c.networkMutex.RLock()
	defer c.networkMutex.RUnlock()
	c.Mutex.RLock()
	defer c.Mutex.RUnlock()
	return SavedNetworkState{
		Network:    c.network,
		MyNode:     c.myNode,
		PrivateKey: c.privateKey,
	}
}

func LoadNetwork(s SavedNetworkState) Cloud {
	utils.GetLogger().Println("[INFO] Loading cloud network.")

	for _, n := range s.Network.Nodes {
		c, err := BootstrapToNetwork(n.IP, s.MyNode, s.PrivateKey)
		if err != nil {
			continue
		}
		return c
	}
	utils.GetLogger().Println("[INFO] Could not reconnect to the network. Starting our own.")
	c := SetupNetwork(s.Network, s.MyNode, s.PrivateKey)
	return c
}
