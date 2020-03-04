package network

import (
	"cloud/datastore"
	"cloud/utils"
	"errors"
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

// Mapping from Node ID's to a slice of Chunk SequenceNumber's.
type distributionScheme map[string][]int

// Distribute computes how to distribute a file and saves the file chunks on the cloud.
// numReplicas specifies how many copies of all file's chunks should be stored on the cloud.
// Note that a replica does not include the original file itself.
// So we store numReplicas+1 contents of the same file on the cloud.
// antiAffinity specifies whether to avoid storing replicas of the same chunk on the same node.
// if numReplicas is -1, then a copy of the file is stored on each node in the cloud.
// Distribute acts with two goals in mind: reliability (redundancy) and efficiency.
// The function uses node benchmarking to achieve best efficiency (load balanced storage, optimized network, etc).
func (c *cloud) Distribute(cloudPath string, file datastore.File, numReplicas int, antiAffinity bool) error {
	cloudPath = CleanNetworkPath(cloudPath)
	// Distribute computes a distributionScheme, a mapping telling which nodes should contain which chunks.
	// It then acts on the distributionScheme to perform the actual requests for saving the chunks.
	distributionScheme, err := c.distributionAlgorithm(file, numReplicas, antiAffinity)
	if err != nil {
		return err
	}
	utils.GetLogger().Printf("[DEBUG] Distribution scheme retrieved: %v.", distributionScheme)

	// Apply the scheme.
	for nodeID, sequenceNumbers := range distributionScheme {
		for _, sequenceNumber := range sequenceNumbers {
			cnode := c.GetCloudNode(nodeID)
			if cnode != nil {
				utils.GetLogger().Printf("[INFO] Saving chunk: %v on node %v.", sequenceNumber, nodeID)
				//err := cnode.SaveChunk(&file, sequenceNumber)
				store := c.FileStore(cloudPath)
				if store == nil {
					return errors.New("file is not stored")
				}
				content, err := store.ReadChunk(file.Chunks.Chunks[sequenceNumber].ID)
				if err != nil {
					return err
				}
				err = cnode.SaveChunk(cloudPath, file.Chunks.Chunks[sequenceNumber], content)
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
	scheme := make(distributionScheme)

	if numReplicas < -1 {
		// TODO: return error?
		return nil, errors.New("numReplicas must be greater than or equal to -1")
	}

	// Get all the nodes we are currently connected to.
	availableNodes := make([]*cloudNode, 0)
	for _, cnode := range c.Nodes {
		availableNodes = append(availableNodes, cnode)
	}

	if len(availableNodes) == 0 {
		// TODO: Might want to replace an error message with a custom error type.
		return nil, errors.New("No nodes available")
	}
	utils.GetLogger().Printf("[DEBUG] Got available nodes: %v.", availableNodes)

	// Get node benchmarks once to not block the network.
	nodeBenchmarks := make([]NodeBenchmark, 0)
	for _, cnode := range availableNodes {
		benchmark, err := cnode.Benchmark()
		if err != nil {
			utils.GetLogger().Printf("[ERROR] %v", err)
			continue
		}
		nodeBenchmarks = append(nodeBenchmarks, benchmark)
	}

	// Apply hard constraints (must be met) on nodes.
	availableNodes, nodeBenchmarks = filterNodes(availableNodes, nodeBenchmarks)
	if len(availableNodes) == 0 {
		return nil, errors.New("No nodes available")
	}
	utils.GetLogger().Printf("[DEBUG] Filtered available nodes: %v.", availableNodes)

	if numReplicas == -1 {
		utils.GetLogger().Printf("[DEBUG] Distributing file to all nodes.")
		allSequenceNumbers := make([]int, 0)
		for i := 0; i < file.Chunks.NumChunks; i++ {
			allSequenceNumbers = append(allSequenceNumbers, i)
		}
		for _, n := range availableNodes {
			scheme[n.ID] = allSequenceNumbers
		}
	}

	// numReplicas is >= 0
	// We use a loop and the modulus operator to iterate over chunks multiple times (creating replicas this way).
	for i := 0; i < file.Chunks.NumChunks*(numReplicas+1); i++ {
		chunk := file.Chunks.Chunks[i%file.Chunks.NumChunks]
		sequenceNumber := chunk.SequenceNumber
		utils.GetLogger().Printf("[DEBUG] Working with Chunk (SequenceNumber): %d.", chunk.SequenceNumber)

		// Apply soft constraints (desired but may not be met) to get the best node.
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

	// FIXME: Don't save a replica multiple times on the same node when anti-affinity is off.
	// Have a count of how many replicas were made, and return that count.
	// See thread at:
	// https://gitlab.computing.dcu.ie/baltrut2/2020-ca326-tbaltrunas-cloudstorage/merge_requests/18#note_12130
	// Implementation detail: Would probably need a "distributionAlgoChunk(chunk, numReplicas)" function.
}

func filterNodes(availableNodes []*cloudNode, benchmarks []NodeBenchmark) ([]*cloudNode, []NodeBenchmark) {
	var newAvailableNodes []*cloudNode
	var newBenchmarks []NodeBenchmark
	for i, n := range availableNodes {
		benchmark := benchmarks[i]

		if benchmark.StorageSpaceRemaining == 0 {
			// node not allowed to store data
			continue
		}
		newAvailableNodes = append(newAvailableNodes, n)
		newBenchmarks = append(newBenchmarks, benchmark)
	}
	return newAvailableNodes, newBenchmarks
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
	// TODO: proper handling of big numbers

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

func expectedOccupation(nodeID string, currentScheme distributionScheme, file datastore.File) uint64 {
	var expectedOccupation uint64 = 0
	seqNums, ok := currentScheme[nodeID]
	if ok {
		for _, seqNum := range seqNums {
			// TODO: method to get chunk by sequence number. Encapsulate file in methods.
			ch := file.Chunks.Chunks[seqNum]
			expectedOccupation += ch.ContentSize
		}
	}
	return expectedOccupation
}
