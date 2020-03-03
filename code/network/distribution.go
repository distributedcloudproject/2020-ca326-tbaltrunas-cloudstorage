package network

import (
	"cloud/datastore"
	"cloud/utils"
)

// Distribute computes how to distribute a chunk and calls the requests.
func (c *cloud) DistributeChunk(cloudPath string, store datastore.FileStore, chunkID datastore.ChunkID) error {
	// TODO: Actual distribution algorithm. For now we copy all chunks to each node.
	cloudPath = CleanNetworkPath(cloudPath)

	content, err := store.ReadChunk(chunkID)
	if err != nil {
		return err
	}
	c.NodesMutex.RLock()
	defer c.NodesMutex.RUnlock()
	for _, n := range c.Nodes {
		utils.GetLogger().Printf("[INFO] Saving chunk: %v on node %v.", chunkID, n.ID)
		chunk, _ := store.Chunk(chunkID)
		if err := n.SaveChunk(cloudPath, chunk, content); err != nil {
			return err
		}
	}
	return nil
}
