package network

import (
	"encoding/gob"
	"io"
)

func (c *Cloud) Save() error {
	if c.SaveFunc != nil {
		encoder := gob.NewEncoder(c.SaveFunc())
		return encoder.Encode(c.Network)
	}
	return nil
}

func (c *Cloud) LoadNetwork (r io.Reader) (*Cloud, error) {
	decoder := gob.NewDecoder(r)
	err := decoder.Decode(&c.Network)
	if err != nil {
		return nil, err
	}

	for i := range c.Network.Nodes {
		c.connectToNode(c.Network.Nodes[i])
	}

	return c, nil
}
