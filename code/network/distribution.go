package network

import (
	"cloud/datastore"
	"cloud/utils"
)

// Mapping from chunk SequenceNumber's to a slice of Node ID's.
type DistributionScheme map[int][]string

// Distribute computes how to distribute a file and saves the chunks on the required nodes.
// It calls DistributionAlgorithm to retrieve a DistributionScheme, a mapping telling which nodes should contain which chunks.
// It then acts on the DistributionScheme to perform the actual requests for saving the chunks.
func Distribute(file *datastore.File, cloud Cloud) error {
	numReplicas := -1
	antiAffinity := true
	// TODO: mutex while reading cloud state
	distributionScheme := DistributionAlgorithm(file, cloud, numReplicas, antiAffinity)
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

// DistributionAlgorithm returns a suitable DistributionScheme for the file and the given cloud.
// It acts with two goals in mind: reliability (redundancy) and efficiency.
// numReplicas specifies how many copies of all file's chunks should be stored on the cloud.
// antiAffinity specifies whether to avoid storing replicas of the same chunk on the same node.
// The function also uses node benchmarking to achieve best efficiency (load balancing storage, and using the fastest network).
func DistributionAlgorithm(file *datastore.File, cloud Cloud, numReplicas int, antiAffinity bool) DistributionScheme {
	if numReplicas == -1 {
		return DistributionAll(file, cloud)
	}
	return nil
}

// DistributionAll specifies to store a copy of the file on each node.
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
