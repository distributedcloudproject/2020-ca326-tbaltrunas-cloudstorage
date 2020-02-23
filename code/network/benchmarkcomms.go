package network

import (
	"cloud/utils"
)

const (
	StorageSpaceRemainingMsg = "StorageSpaceRemaining"
)

func init() {
	handlers = append(handlers, createBenchmarkRequestHandler)
}

// TODO: display bytes in terms of megabytes (divide by 1024^2)
func (n *cloudNode) StorageSpaceRemaining() (int64, error) {
	ret, err := n.client.SendMessage(StorageSpaceRemainingMsg)
	return ret[0].(int64), err
}

func (r request) OnStorageSpaceRemaining() (int64, error) {
	r.Cloud.Mutex.Lock()
	defer r.Cloud.Mutex.Unlock()

	storageDir := r.Cloud.config.FileStorageDir
	utils.GetLogger().Printf("[DEBUG] Storage Path on the node: %s.", storageDir)
	storageCapacity := r.Cloud.config.FileStorageCapacity
	utils.GetLogger().Printf("[DEBUG] Storage Capacity on the node: %d.", storageCapacity)
	if storageCapacity == 0 {
		// StorageCapacity not set, calculate available disk space
		spaceRemaining := utils.AvailableDisk(storageDir)
		return spaceRemaining, nil
	} else {
		// Walk through the directory to calculate current usage.
		spaceUsed, err := utils.DirSize(storageDir)
		if err != nil {
			return 0, err
		}
		utils.GetLogger().Printf("[DEBUG] Computed space usage: %d.", spaceUsed)
		spaceRemaining := storageCapacity - spaceUsed
		return spaceRemaining, nil
	}
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
		}
		return nil
	}
}