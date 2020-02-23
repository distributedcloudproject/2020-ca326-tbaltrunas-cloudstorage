package network

import (
	"cloud/datastore"
	"cloud/utils"
	"errors"
)

// Mapping from Node ID's to a slice of Chunk SequenceNumber's.
type distributionScheme map[string][]int

// Distribute computes how to distribute a file and saves the file chunks on the cloud..
// numReplicas specifies how many copies of all file's chunks should be stored on the cloud.
// antiAffinity specifies whether to avoid storing replicas of the same chunk on the same node.
// if numReplicas is -1, then a copy of the file is stored on each node in the cloud.
// Distribute acts with two goals in mind: reliability (redundancy) and efficiency.
// The function uses node benchmarking to achieve best efficiency (load balanced storage, optimized network, etc).
func (c *cloud) Distribute(file datastore.File, numReplicas int, antiAffinity bool) error {
	// TODO: mutex while reading cloud state

	// Internally Distribute computes a distributionScheme, a mapping telling which nodes should contain which chunks.
	// It then acts on the distributionScheme to perform the actual requests for saving the chunks.
	distributionScheme, err := c.distributionAlgorithm(file, numReplicas, antiAffinity)
	if err != nil {
		return err
	}
	utils.GetLogger().Printf("[DEBUG] Distribution scheme retrieved: %v.", distributionScheme)

	for nodeID, sequenceNumbers := range distributionScheme {
		for _, sequenceNumber := range sequenceNumbers {
			cnode := c.GetCloudNode(nodeID)
			if cnode != nil {
				utils.GetLogger().Printf("[INFO] Saving chunk: %v on node %v.", sequenceNumber, nodeID)
				err := cnode.SaveChunk(&file, sequenceNumber)
				if err != nil {
					return err
				}
				// TODO: re-run operation in case of error.
			}
		}
	}
	return nil
}

// distributionAlgorithm returns a suitable distributionScheme for the file and the given cloud.
func (c *cloud) distributionAlgorithm(file datastore.File, numReplicas int, antiAffinity bool) (distributionScheme, error) {
	// FIXME: pass fileID instead of file struct? Then reference the datastore?
	if numReplicas == -1 {
		return c.distributionAll(file)
	} else if numReplicas < -1 {
		// TODO: return error?
		return nil, errors.New("numReplicas must be greater than or equal to -1.")
	}

	// get benchmarks once, then reuse for each chunk (to not block the network mutex)
	availableNodes := c.GetAllCloudNodes()
	// TODO: check that availableNodes is len > 0

	utils.GetLogger().Printf("[DEBUG] Got available nodes: %v.", availableNodes)
	nodeBenchmarks := make([]NodeBenchmark, 0)
	for _, cnode := range availableNodes {
		benchmark, err := cnode.Benchmark()
		if err != nil {
			utils.GetLogger().Printf("[ERROR] %v", err)
			continue
		}
		nodeBenchmarks = append(nodeBenchmarks, benchmark)
	}
	scheme := make(distributionScheme)
	// We use a loop and the modulus operator to iterate over chunks multiple times (creating replicas this way).
	for i := 0; i < file.Chunks.NumChunks * (numReplicas + 1); i++ {
		chunk := file.Chunks.Chunks[i % file.Chunks.NumChunks]
		sequenceNumber := chunk.SequenceNumber
		utils.GetLogger().Printf("[DEBUG] Working with Chunk (SequenceNumber): %d.", chunk.SequenceNumber)

		chosenNode, err := c.bestNode(availableNodes, scheme, sequenceNumber, file, antiAffinity, nodeBenchmarks)
		if err != nil {
			return nil, err
		}
		utils.GetLogger().Printf("[DEBUG] Chosen node to distribute chunk to: %s.", chosenNode.ID)

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

// distributionAll specifies to store a copy of the file on each node.
func (c *cloud) distributionAll(file datastore.File) (distributionScheme, error) {
	utils.GetLogger().Printf("[DEBUG] Distributing file to all nodes.")
	scheme := make(distributionScheme)
	allSequenceNumbers := make([]int, 0)
	for i := 0; i < file.Chunks.NumChunks; i++ {
		allSequenceNumbers = append(allSequenceNumbers, i)
	}
	for _, n := range c.Network().Nodes {
		scheme[n.ID] = allSequenceNumbers
	}
	return scheme, nil
}

func (c *cloud) bestNode(availableNodes []*cloudNode, currentScheme distributionScheme, chunkSequenceNumber int, 
						 file datastore.File, antiAffinity bool, benchmarks []NodeBenchmark) (*cloudNode, error) {
		scores := make([]int, 0)
		for i, n := range availableNodes {
			score, err := c.Score(n, benchmarks[i], currentScheme, chunkSequenceNumber, file, antiAffinity)
			if err != nil {
				return nil, err
			}
			scores = append(scores, score)
		}
		utils.GetLogger().Printf("[DEBUG] Got scores for each node: %v.", scores)

		_, idx, err := utils.MaxInt(scores)
		if err != nil {
			return nil, err
		}
		return availableNodes[idx], nil
}

func (c *cloud) Score(cnode *cloudNode, benchmark NodeBenchmark, currentScheme distributionScheme, 
					  chunkSequenceNumber int, file datastore.File, antiAffinity bool) (int, error) {
	score := 0
	
	storageSpaceRemaining := benchmark.StorageSpaceRemaining
	// Note that we do not distribute the current chunks until the distributionScheme is fully constructed.
	expectedOccupation := expectedOccupation(cnode.ID, currentScheme, file)
	expectedStorageRemaining := storageSpaceRemaining - expectedOccupation
	score += int(expectedStorageRemaining)

	if antiAffinity {
		antiAffine := upholdsAntiAffinity(cnode.ID, chunkSequenceNumber, currentScheme) // does not contain the chunk already
		score += 1
		if !antiAffine {
			score *= -1
		}
	}
	return score, nil
}

func upholdsAntiAffinity(nodeID string, chunkSequenceNumber int, currentScheme distributionScheme) bool {
	seqNums, ok := currentScheme[nodeID]
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

func expectedOccupation(nodeID string, currentScheme distributionScheme, file datastore.File) int64 {
	var expectedOccupation int64
	seqNums, ok := currentScheme[nodeID]
	if ok {
		for _, seqNum := range seqNums {
			// TODO: method to get chunk by sequence number. Encapsulate file in methods.
			ch := file.Chunks.Chunks[seqNum]
			expectedOccupation += int64(ch.ContentSize)  // FIXME: mixing int64 and int types - just stick to one
		}
	}
	return expectedOccupation
}
