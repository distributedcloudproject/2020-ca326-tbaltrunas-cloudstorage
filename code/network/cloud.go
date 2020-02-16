package network

import (
	"cloud/comm"
	"errors"
	"cloud/utils"
)

func (c *Cloud) connectToNode(n *Node) error {
	utils.GetLogger().Printf("[INFO] Cloud: %v, connecting to node: %v", c, n)
	n.mutex.Lock()
	defer n.mutex.Unlock()
	if n.IP != "" && n.ID != c.MyNode.ID && n.client == nil {
		utils.GetLogger().Printf("[DEBUG] Connecting to a non-me node with nil client: %v.", n)
		var err error
		n.client, err = comm.NewClientDial(n.IP, c.PrivateKey)
		utils.GetLogger().Printf("[DEBUG] Node with added client: %v.", n)
		if err != nil {
			return err
		}
		n.client.AddRequestHandler(createAuthRequestHandler(n, c))
		go n.client.HandleConnection()
		n.Authenticate(c.MyNode)
		n.client.AddRequestHandler(createRequestHandler(n, c))
	}else if n.ID == c.MyNode.ID && n.client == nil {
		n.client = comm.NewLocalClient()
	}
	return nil
}

func (c *Cloud) addNode(node *Node) {
	utils.GetLogger().Printf("[INFO] Adding node to cloud: %v.", node)
	c.NodeMutex.Lock()
	defer c.NodeMutex.Unlock()

	for _, n := range c.Network.Nodes {
		if n.ID == node.ID {
			if n.client == nil {
				utils.GetLogger().Printf("[DEBUG] Found matching node with nil client: %v.", n)
				n.IP = node.IP
				n.Name = node.Name
				n.client = node.client
				utils.GetLogger().Printf("[DEBUG] Updated matching nil client node: %v.", n)
			}
			return
		}
	}
	c.Network.Nodes = append(c.Network.Nodes, node)
	c.Save()
}

func (c *Cloud) OnlineNodesNum() int {
	utils.GetLogger().Println("[DEBUG] Getting the number of nodes online.")
	c.NodeMutex.RLock()
	defer c.NodeMutex.RUnlock()

	i := 0
	for _, n := range c.Network.Nodes {
		if n.client != nil || n.ID == c.MyNode.ID {
			utils.GetLogger().Printf("[DEBUG] Node with non-nil client or a me-node: %v.", n)
			i++
		}
	}
	utils.GetLogger().Printf("[DEBUG] Number of online nodes counted: %v.", i)
	return i
}

func (c *Cloud) addToWhitelist(ID string) error {
	if ID == "" {
		return errors.New("cannot add empty ID")
	}

	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	for i := range c.Network.WhitelistIDs {
		if c.Network.WhitelistIDs[i] == ID {
			return errors.New("ID is already whitelisted")
		}
	}
	c.Network.WhitelistIDs = append(c.Network.WhitelistIDs, ID)

	return nil
}

func (c *Cloud) AddToWhitelist(ID string) error {
	err := c.addToWhitelist(ID)
	if err != nil {
		return err
	}

	c.NodeMutex.RLock()
	defer c.NodeMutex.RUnlock()
	for _, n := range c.Network.Nodes {
		if n.client != nil && n.ID != c.MyNode.ID {
			n.AddToWhitelist(ID)
		}
	}
	return nil
}

func (c *Cloud) IsWhitelisted(ID string) bool {
	if ID == "" {
		return false
	}

	c.Mutex.RLock()
	defer c.Mutex.RUnlock()

	for i := range c.Network.WhitelistIDs {
		if c.Network.WhitelistIDs[i] == ID {
			return true
		}
	}
	return false
}