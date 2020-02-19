package network

import (
	"cloud/datastore"
	"cloud/utils"
)

// Distribute computes how to distribute a file and calls the requests.
func Distribute(file *datastore.File, cloud Cloud) error {
	// TODO: Actual distribution algorithm. For now we copy all chunks to each node.
	nodes := cloud.Network().Nodes
	for _, n := range nodes {
		// TODO: check if n.client != nil
		// Maybe distribution should be moved to network package?
		for i, _ := range file.Chunks.Chunks {
			utils.GetLogger().Printf("[INFO] Saving chunk: %v on node %v.", i, n.Name)
			cnode := cloud.GetCloudNode(n.ID)
			if cnode != nil {
				err := cnode.SaveChunk(file, i)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
