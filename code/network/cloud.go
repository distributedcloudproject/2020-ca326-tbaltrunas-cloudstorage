package network

import (
	"cloud/comm"
	"cloud/datastore"
	"cloud/utils"
	"errors"
	"os"
	"path/filepath"
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
		n.client.AddRequestHandler(createDataStoreRequestHandler(n, c))
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

func (c *Cloud) addFile(file *datastore.File) error {
	utils.GetLogger().Printf("[INFO] Adding file to cloud: %v.", file)

	// check if file is not already added
	c.Mutex.RLock()
	ok := c.Network.DataStore.Contains(file)
	c.Mutex.RUnlock()
	if ok {
		// exit preemptively
		// because if the node already has the file, then all other nodes should also have it
		return nil
	}

	// add file
	c.Mutex.Lock()
	c.Network.DataStore.Add(file)
	err := c.Save()
	c.Mutex.Unlock()
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

// saveChunk persistently stores a chunk given by its contents, as the given cloud path.
func (c *Cloud) saveChunk(cloudPath string, chunk datastore.Chunk, contents []byte) error {
	// TODO: verify chunk ID
	// chunkID := chunk.ID

	// persistently store chunk
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	
	localPath := filepath.Join(c.MyNode.FileStorageDir, cloudPath)
	utils.GetLogger().Printf("[DEBUG] Computed path where to store chunk: %s.", localPath)
	// TODO: maybe this should be done when setting up the node
	err := os.MkdirAll(filepath.Dir(localPath), os.ModeDir)
	if err != nil {
		return err
	}
	utils.GetLogger().Println("[DEBUG] Created/verified existence of path directories.")
	w, err := os.Create(localPath)
	if err != nil {
		return err 
	}
	defer w.Close()

	utils.GetLogger().Printf("[DEBUG] Saving chunk to writer: %v.", w)
	err = datastore.SaveChunk(w, contents)
	if err != nil { 
		return err
	}
	utils.GetLogger().Printf("[DEBUG] Finished saving chunk to writer: %v.", w)
	return nil
}

func (c *Cloud) updateFileChunkLocations(chunkID datastore.ChunkID, nodeID string) {
	utils.GetLogger().Printf("[DEBUG] Updating FileChunkLocations with ChunkID: %v, NodeID: %v.", 
							  chunkID, nodeID)

	c.Mutex.Lock()

	chunkNodes, ok := c.Network.FileChunkLocations[chunkID]
	if ok {
		utils.GetLogger().Printf("[DEBUG] Got list of nodes for key (%v): %v.", chunkID, chunkNodes)
		for _, nID := range chunkNodes {
			if nID == nodeID {
				// node already added
				// pre-emptive exit.
				utils.GetLogger().Println("[DEBUG] FileChunkLocations already contains the needed key-value.")
				c.Mutex.Unlock()

				return
			}
		}
	}

	// update our own chunk location data structure
	utils.GetLogger().Printf("[DEBUG] Updating FileChunkLocations: %v.", c.Network.FileChunkLocations)
	if !ok {
		// key not present
		utils.GetLogger().Printf("[DEBUG] Creating a new list for the key: %v, in FileChunkLocations.", chunkID)
		chunkNodes = []string{nodeID}
	} else {
		// key present and has other nodes
		utils.GetLogger().Printf("[DEBUG] Appending to the list for the key: %v, in FileChunkLocations.", chunkID)
		chunkNodes = append(chunkNodes, nodeID)
	}
	c.Network.FileChunkLocations[chunkID] = chunkNodes
	utils.GetLogger().Printf("[DEBUG] Finished updating FileChunkLocations: %v.", c.Network.FileChunkLocations)

	c.Mutex.Unlock()

	// propagate change to other nodes
	utils.GetLogger().Println("[DEBUG] Communicating change in FileChunkLocations to other nodes.")
	for _, n := range c.Network.Nodes {
		if n.client != nil && n.ID != c.MyNode.ID {
			utils.GetLogger().Printf("[DEBUG] Found node to communicate change in FileChunkLocations to: %v.", 
									  n)
			n.updateFileChunkLocations(chunkID, nodeID)
		}
	}
	utils.GetLogger().Println("[DEBUG] Finished communicating changes in FileChunkLocations to other nodes.")
}
