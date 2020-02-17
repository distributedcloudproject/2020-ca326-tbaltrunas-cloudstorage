package network

import (
	"cloud/comm"
	"cloud/datastore"
	"cloud/utils"
)

func (c *Cloud) connectToNode(n *Node) error {
	utils.GetLogger().Printf("[INFO] Cloud: %v, connecting to node: %v", c, n)
	n.mutex.Lock()
	defer n.mutex.Unlock()
	if n.IP != "" && n.ID != c.MyNode.ID && n.client == nil {
		utils.GetLogger().Printf("[DEBUG] Connecting to a non-me node with nil client: %v.", n)
		var err error
		n.client, err = comm.NewClientDial(n.IP)
		utils.GetLogger().Printf("[DEBUG] Node with added client: %v.", n)
		if err != nil {
			return err
		}
		n.client.AddRequestHandler(createAuthRequestHandler(n, c))
		go n.client.HandleConnection()
		_ = n.Authenticate(c.MyNode)
		n.client.AddRequestHandler(createRequestHandler(n, c))
		n.client.AddRequestHandler(createDataStoreRequestHandler(n, c))
	}else if n.ID == c.MyNode.ID && n.client == nil {
		n.client = comm.NewLocalClient()
	}
	return nil
}

func (c *Cloud) addNode(node *Node) {
	utils.GetLogger().Printf("[INFO] Adding node to cloud: %v.", node)
	// c.Mutex.Lock()
	// defer c.Mutex.Unlock()

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
	c.Mutex.RLock()
	defer c.Mutex.RUnlock()
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

func (c *Cloud) addFile(file *datastore.File) error {
	utils.GetLogger().Printf("[INFO] Adding file to cloud: %v.", file)

	// FIXME: problem with mutexes
	// c.Mutex.Lock()
	// defer c.Mutex.Unlock()
	// check if file is not already added
	ok := c.Network.DataStore.Contains(file)
	if ok {
		// exit preemptively
		// because if the node already has the file, then all other nodes should also have it
		return nil
	}
	// add file
	c.Network.DataStore.Add(file)
	err := c.Save()
	if err != nil {
		return err
	}

	// repeat command to all other nodes
	utils.GetLogger().Printf("[DEBUG] Network has nodes: %v.", c.Network.Nodes)
	for _, n := range c.Network.Nodes {
		if n.client != nil && n.ID != c.MyNode.ID {
			utils.GetLogger().Printf("[DEBUG] Found non-self node with non-nil client: %v.", n)
			err := n.AddFile(file)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Contains returns whether the datastore contains the specified file.
func (ds *DataStore) Contains(file *datastore.File) bool {
	for _, f := range ds.Files {
		// TODO: file ID
		if file.Path == f.Path {
			return true
		}
	}
	return false
}

// Add appends a file to the datastore.
func (ds *DataStore) Add(file *datastore.File) {
	ds.Files = append(ds.Files, file)
}

// TODO: move DataStore to datastore package
// Also DataStore is just a wrapper for a slice?
