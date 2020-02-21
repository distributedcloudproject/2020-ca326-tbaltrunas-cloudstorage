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

func (n *cloudNode) StorageSpaceRemaining() (int64, error) {
	ret, err := n.client.SendMessage(StorageSpaceRemainingMsg)
	return ret[0].(int64), err
}

func (r request) OnStorageSpaceRemaining() (int64, error) {
	c := r.Cloud

	storageCapacity := c.config.FileStorageCapacity
	utils.GetLogger().Printf("[DEBUG] Storage Capacity on the node: %d.", storageCapacity)
	if storageCapacity == 0 {
		// TODO: If maximum capacity 0, calculate available disk space
		return 0, nil
	} else {
		// Walk through directory
		storageDir := c.config.FileStorageDir
		utils.GetLogger().Printf("[DEBUG] Storage Path on the node: %s.", storageDir)

		spaceUsed, err := utils.DirSize(storageDir)
		if err != nil {
			return 0, err
		}
		utils.GetLogger().Printf("[DEBUG] Computed space usage: %d.", spaceUsed)

		// Subtract contents from maximum capacity.
		// FIXME: validate that capacity > usage ?
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