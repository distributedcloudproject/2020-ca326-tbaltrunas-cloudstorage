package network

import (
	"cloud/utils"
	"encoding/gob"
	"io"
)

func (c *Cloud) Save() error {
	utils.GetLogger().Println("Saving cloud network.")
	if c.SaveFunc != nil {
		utils.GetLogger().Println("Saving cloud network with SaveFunc.")
		encoder := gob.NewEncoder(c.SaveFunc())
		return encoder.Encode(c.Network)
	}
	return nil
}

// LoadNetwork reads the saved network state from the reader and resumes the network.
func (c *Cloud) LoadNetwork (r io.Reader) (error) {
	utils.GetLogger().Println("Loading cloud network.")
	decoder := gob.NewDecoder(r)
	err := decoder.Decode(&c.Network)
	if err != nil {
		return err
	}

	utils.GetLogger().Println("Connecting to each node (resetting client).")
	for i := range c.Network.Nodes {
		c.Network.Nodes[i].client = nil
		c.connectToNode(c.Network.Nodes[i])
	}

	return nil
}
