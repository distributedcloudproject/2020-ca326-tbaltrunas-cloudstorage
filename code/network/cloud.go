package network

import (
	"cloud/comm"
	"cloud/datastore"
	"cloud/utils"
	"os"
	"strconv"
	"path/filepath"
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

func (c *Cloud) saveChunk(path string, chunk datastore.Chunk, contents []byte) error {
	// TODO: verify chunk ID
	chunkID := chunk.ID

	// persistently store chunk
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	
	chunkPath := filepath.Join(c.MyNode.FileStorageDir, filepath.Dir(path), 
								filepath.Base(path) + "-" + strconv.Itoa(chunk.SequenceNumber))
	utils.GetLogger().Printf("[DEBUG] Computed path where to store chunk: %s.", chunkPath)
	// TODO: maybe this should be done when setting up the node
	err := os.MkdirAll(filepath.Dir(chunkPath), os.ModeDir)
	if err != nil {
		return err
	}
	utils.GetLogger().Println("[DEBUG] Created/verified existence of path directories.")
	w, err := os.Create(chunkPath)
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

	// update chunk data structure
	utils.GetLogger().Println("[DEBUG] Updating FileChunkLocations.")
	nodeID := c.MyNode.ID
	chunkLocations, ok := c.Network.FileChunkLocations[chunkID]
	if !ok {
		chunkLocations = []string{nodeID}
	} else {
		chunkLocations = append(chunkLocations, nodeID)
	}
	c.Network.FileChunkLocations[chunkID] = chunkLocations
	utils.GetLogger().Println("[DEBUG] Finished updating FileChunkLocations.")

	// propagate change to other nodes
	utils.GetLogger().Println("[DEBUG] Communicating change in FileChunkLocations to other nodes.")
	for _, n := range c.Network.Nodes {
		if n.client != nil && n.ID != c.MyNode.ID {
			utils.GetLogger().Printf("[DEBUG] Found non-self node with non-nil client: %v.", n)
			err := n.updateFileChunkLocations(c.Network.FileChunkLocations)
			if err != nil {
				return err
			}
		}
	}
	utils.GetLogger().Println("[DEBUG] Finished communicating changes in FileChunkLocations to other nodes.")
	return nil
}
