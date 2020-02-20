package network

import (
	"cloud/datastore"
	"cloud/utils"
)

// Mapping from chunk SequenceNumber's to a slice of Node ID's.
type DistributionScheme map[int][]string

// Distribute computes how to distribute a file and saves the chunks on the required nodes.
func Distribute(file *datastore.File, cloud Cloud) error {
	// TODO: Actual distribution algorithm. For now we copy all chunks to each node.
	nodes := cloud.Network().Nodes
	for _, n := range nodes {
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

func DistributionAlgorithm(file *datastore.File, cloud Cloud, redundancyFactor int, antiAffinity bool) DistributionScheme {
	return nil
}
