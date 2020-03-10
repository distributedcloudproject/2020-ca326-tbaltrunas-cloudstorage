package network

import (
	"cloud/utils"
	"time"
)

// Messages used for benchmark communications.
const (
	StorageSpaceRemainingMsg = "StorageSpaceRemaining"
	NetworkLatencyMsg        = "NetworkLatency"
)

func init() {
	handlers = append(handlers, createBenchmarkRequestHandler)
}

func createBenchmarkRequestHandler(node *cloudNode, cloud *cloud) func(string) interface{} {
	r := request{
		Cloud:    cloud,
		FromNode: node,
	}

	return func(message string) interface{} {
		switch message {
		case StorageSpaceRemainingMsg:
			return r.OnStorageSpaceRemaining
		case NetworkLatencyMsg:
			return r.OnNetworkLatency
		}
		return nil
	}
}

// StorageSpaceRemaining returns the amount of storage space in bytes that remains for user data on the node.
// Returns 0 if node can not store any data.
func (n *cloudNode) StorageSpaceRemaining() (uint64, error) {
	ret, err := n.client.SendMessage(StorageSpaceRemainingMsg)
	return ret[0].(uint64), err
}

func (r request) OnStorageSpaceRemaining() (uint64, error) {
	r.Cloud.Mutex.RLock()
	defer r.Cloud.Mutex.RUnlock()

	storageDir := r.Cloud.config.FileStorageDir
	utils.GetLogger().Printf("[DEBUG] Storage Path on the node: %s.", storageDir)
	storageCapacity := r.Cloud.config.FileStorageCapacity
	utils.GetLogger().Printf("[DEBUG] Storage Capacity on the node: %d.", storageCapacity)
	if storageCapacity == -1 {
		return 0, nil
	} else if storageCapacity == 0 {
		// StorageCapacity not set, calculate available disk space
		spaceRemaining := utils.AvailableDisk(storageDir)
		return spaceRemaining, nil
	} else {
		// Get the current usage from benchmark state.
		spaceUsed := r.Cloud.BenchmarkState().StorageSpaceUsed

		// TODO: An alternative method when no data is available
		// may be to walk through the entire directory to calculate its usage.
		// We can use utils.DirSize for that.

		utils.GetLogger().Printf("[DEBUG] Computed space usage: %d.", spaceUsed)
		spaceRemaining := uint64(storageCapacity) - spaceUsed
		return spaceRemaining, nil
	}
}

// NetworkLatency calculates the round-trip time of a request.
func (n *cloudNode) NetworkLatency() (time.Duration, error) {
	var latency time.Duration
	before := time.Now()
	_, err := n.client.SendMessage(NetworkLatencyMsg)
	latency = time.Since(before)
	return latency, err
}

func (r request) OnNetworkLatency() {}

// TODO: network bandwidth
