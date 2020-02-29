package network

import (
	"cloud/datastore"
	"cloud/utils"
)

// Distribute computes how to distribute a chunk and calls the requests.
func (c *cloud) DistributeChunk(chunk datastore.ChunkStore) error {
	// TODO: Actual distribution algorithm. For now we copy all chunks to each node.

	content, err := c.readChunk(chunk)
	if err != nil {
		return err
	}
	c.NodesMutex.RLock()
	defer c.NodesMutex.RUnlock()
	for _, n := range c.Nodes {
		utils.GetLogger().Printf("[INFO] Saving chunk: %v on node %v.", chunk.Chunk.ID, n.ID)
		if err := n.SaveChunk(chunk.FileID, chunk.Chunk, content); err != nil {
			return err
		}
	}
	return nil
}
