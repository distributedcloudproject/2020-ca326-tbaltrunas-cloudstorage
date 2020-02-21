package network

import (
	"cloud/datastore"
	"cloud/utils"
	"errors"
)

// Mapping from Node ID's to a slice of Chunk SequenceNumber's.
// TODO: maybe other way around is better. Nodes are unique. Chunks may be replicated!
type DistributionScheme map[string][]int

// Distribute computes how to distribute a file and saves the chunks on the required nodes.
// It calls DistributionAlgorithm to retrieve a DistributionScheme, a mapping telling which nodes should contain which chunks.
// It then acts on the DistributionScheme to perform the actual requests for saving the chunks.
func Distribute(file *datastore.File, cloud Cloud, numReplicas int, antiAffinity bool) error {
	// TODO: mutex while reading cloud state
	distributionScheme, err := DistributionAlgorithm(file, cloud, numReplicas, antiAffinity)
	if err != nil {
		return err
	}
	utils.GetLogger().Printf("[DEBUG] Distribution scheme retrieved: %v.", distributionScheme)
	// TODO: Actual distribution algorithm. For now we copy all chunks to each node.

	for nodeID, sequenceNumbers := range distributionScheme {
		for _, sequenceNumber := range sequenceNumbers {
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
func DistributionAlgorithm(file *datastore.File, cloud Cloud, numReplicas int, antiAffinity bool) (DistributionScheme, error) {
	if numReplicas == -1 {
		return DistributionAll(file, cloud)
	} else if numReplicas < -1 {
		// TODO: return error?
		return nil, errors.New("numReplicas must be greater than or equal to -1.")
	}

	availableNodes := cloud.Network().Nodes // TODO: sort by benchmark - need to re-run each time, i.e. include chunk in storage calc?
	scheme := make(DistributionScheme)
	// We use a loop and the modulus operator to iterate over chunks multiple times (creating replicas this way).
	for i := 0; i < file.Chunks.NumChunks * (numReplicas + 1); i++ {
		chunk := file.Chunks.Chunks[i % file.Chunks.NumChunks]
		sequenceNumber := chunk.SequenceNumber
		utils.GetLogger().Printf("[DEBUG] Working with Chunk (SequenceNumber): %d.", chunk.SequenceNumber)
		// FIXME: need to replace Network().Nodes with cloudNodes (use only nodes you can connect to).
		// sorted and filtered priority list
		// candidateNodes := applyAffinityRule(sequenceNumber, availableNodes, scheme, antiAffinity) // TODO: filter and sort
		// TODO: check that candidateNodes / allNodes is len > 0
		scores := make([]int, 0)
		for _, n := range availableNodes {
			score := 0
			
			if antiAffinity {
				affine := isAffine(n, sequenceNumber, scheme) // does not contain the chunk already
				if affine {
					score += 10
					score *= 2
				}
			}

			// StorageRemaining returns the amount of storage remaining on a node.
			// func StorageRemaining() int
			// storageSpaceRemaining := StorageRemaining(n)
			cnode := cloud.GetCloudNode(n.ID)
			if cnode != nil {
				storageSpaceRemaining, err := cnode.StorageSpaceRemaining()
				if err != nil {
					return nil, err
				}
				utils.GetLogger().Printf("[DEBUG] Storage space remaining: %d.", storageSpaceRemaining)
				// Note that we do not distribute the current chunks until the scheme is fully constructed.

				var expectedOccupation int64
				seqNums, ok := scheme[n.ID]
				if ok {
					for _, seqNum := range seqNums {
						// TODO: method to get chunk by sequence number. Encapsulate file in methods.
						ch := file.Chunks.Chunks[seqNum]
						expectedOccupation += int64(ch.ContentSize)  // FIXME: mixing int64 and int types - just stick to one
					}
				}
				expectedStorageRemaining := storageSpaceRemaining - expectedOccupation
				score += int(expectedStorageRemaining)
			}
			// FIXME: if nil, remove from list of availableNodes.


			scores = append(scores, score)
		}
		utils.GetLogger().Printf("[DEBUG] Got scores for each node: %v.", scores)

		bestScore := 0
		bestScoreIdx := 0
		for i, score := range scores {
			if bestScore < score {
				bestScore = score
				bestScoreIdx = i
			}
		}
		chosenNode := availableNodes[bestScoreIdx] // we will send the chunk on this node.
		utils.GetLogger().Printf("[DEBUG] Chosen node to distribute chunk to (Node name): %s.", chosenNode.Name)

		// TODO: move this into a method
		nodeID := chosenNode.ID
		sequenceNumbers, ok := scheme[nodeID]
		if !ok {
			sequenceNumbers = []int{sequenceNumber}
		} else {
			sequenceNumbers = append(sequenceNumbers, sequenceNumber)
		}
		scheme[nodeID] = sequenceNumbers
	}
	return scheme, nil
}

// DistributionAll specifies to store a copy of the file on each node.
func DistributionAll(file *datastore.File, cloud Cloud) (DistributionScheme, error) {
	scheme := make(DistributionScheme)
	allSequenceNumbers := make([]int, 0)
	for i := 0; i < file.Chunks.NumChunks; i++ {
		allSequenceNumbers = append(allSequenceNumbers, i)
	}
	for _, n := range cloud.Network().Nodes {
		scheme[n.ID] = allSequenceNumbers
	}
	return scheme, nil
}

// FIXME: make functions private.

func isAffine(n Node, chunkSequenceNumber int, currentScheme DistributionScheme) bool {
	seqNums, ok := currentScheme[n.ID]
	if !ok {
		return true
	}
	for _, seqNum := range seqNums {
		if seqNum == chunkSequenceNumber {
			// chunk is already on the node
			return false
		}
	}
	return true
}
