package network

import (
	"cloud/datastore"
	"cloud/utils"
)

// Mapping from chunk SequenceNumber's to a slice of Node ID's.
type DistributionScheme map[int][]string

// Distribute computes how to distribute a file and saves the chunks on the required nodes.
func Distribute(file *datastore.File, cloud Cloud) error {
	redundancyFactor := -1
	antiAffinity := true
	// TODO: mutex while reading cloud state
	distributionScheme := DistributionAlgorithm(file, cloud, redundancyFactor, antiAffinity)
	// TODO: Actual distribution algorithm. For now we copy all chunks to each node.

	for sequenceNumber, nodeIDs := range distributionScheme {
		for _, nodeID := range nodeIDs {
			cnode := cloud.GetCloudNode(nodeID)
			if cnode != nil {
				utils.GetLogger().Printf("[INFO] Saving chunk: %v on node %v.", sequenceNumber, nodeID)
				err := cnode.SaveChunk(file, sequenceNumber)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func DistributionAlgorithm(file *datastore.File, cloud Cloud, redundancyFactor int, antiAffinity bool) DistributionScheme {
	if redundancyFactor == -1 {
		return DistributionAll(file, cloud)
	} else {
		return nil
	}
}

func DistributionAll(file *datastore.File, cloud Cloud) DistributionScheme {
	scheme := make(map[int][]string)
	allNodeIDs := make([]string, 0)
	for _, n := range cloud.Network().Nodes {
		allNodeIDs = append(allNodeIDs, n.ID)
	}
	for i, _ := range file.Chunks.Chunks {
		scheme[i] = allNodeIDs
	}
	return scheme
}
