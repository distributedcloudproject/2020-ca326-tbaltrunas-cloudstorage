package network

import (
	"cloud/comm"
)

func (c *Cloud) connectToNode(n *Node) error {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	if n.IP != "" && n.ID != c.MyNode.ID && n.client == nil {
		var err error
		n.client, err = comm.NewClientDial(n.IP)
		if err != nil {
			return err
		}
		n.client.AddRequestHandler(createAuthRequestHandler(n, c))
		go n.client.HandleConnection()
		_ = n.Authenticate(c.MyNode)
		n.client.AddRequestHandler(createRequestHandler(n, c))
	}else if n.ID == c.MyNode.ID && n.client == nil {
		n.client = comm.NewLocalClient()
	}
	return nil
}

func (c *Cloud) addNode(node *Node) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()

	for _, n := range c.Network.Nodes {
		if n.ID == node.ID {
			if n.client == nil {
				n.IP = node.IP
				n.Name = node.Name
				n.client = node.client
			}
			return
		}
	}
	c.Network.Nodes = append(c.Network.Nodes, node)
	c.Save()
}

func (c *Cloud) OnlineNodesNum() int {
	c.Mutex.RLock()
	defer c.Mutex.RUnlock()
	i := 0
	for _, n := range c.Network.Nodes {
		if n.client != nil || n.ID == c.MyNode.ID {
			i++
		}
	}
	return i
}