package distribution

import (
	"cloud/datastore"
	"cloud/network"
	"cloud/utils"
)

// Distribute computes how to distribute a file and calls the requests.
func Distribute(file *datastore.File, cloud *network.Cloud) error {
	// TODO: Actual distribution algorithm. For now we copy all chunks to each node.
	for _, n := range cloud.Network.Nodes {
		// TODO: check if n.client != nil
		// Maybe distribution should be moved to network package?
		for i, _ := range file.Chunks.Chunks {
			utils.GetLogger().Printf("[INFO] Saving chunk: %v on node %v.", i, n.Name)
			err := n.SaveChunk(file, i)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
