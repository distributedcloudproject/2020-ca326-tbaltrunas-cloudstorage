package network

import (
	"cloud/datastore"
	"cloud/utils"
	"testing"
)

func TestChunkDistribution(t *testing.T) {
	numNodes := 5
	clouds, err := CreateTestClouds(numNodes)

	tmpStorageDirs, err := utils.GetTestDirs("cloud_test_node_data_", numNodes)
	if err != nil {
		t.Fatal(err)
	}
	defer utils.GetTestDirsCleanup(tmpStorageDirs)
	// TODO: create multiple test cases with different clouds and expected distributions.
	// Test cases:
	// - Storage (same)
	// - Storage (different)
	// - Storage (unlimited)
	// - Affinity
	storageCapacities := []int64{100, 100, 100, 100, 100}
	for i, cloud := range clouds {
		cloud.SetConfig(CloudConfig{
			FileStorageDir: tmpStorageDirs[i],
			FileStorageCapacity: storageCapacities[i],
		})
	}

	t.Logf("Test clouds: %v.", clouds)
	t.Logf("Storage locations for clouds: %v.", tmpStorageDirs)

	cloud := clouds[0]
	t.Logf("Main cloud: %v.", cloud)
	t.Logf("MyNode on cloud: %v.", cloud.MyNode())
	t.Logf("Cloud with other nodes: %v.", cloud)
	t.Logf("Network: %v.", cloud.Network())
	nodes := cloud.Network().Nodes
	for i := range nodes {
		t.Logf("Node %d: %v.", i, nodes[i])
	}

	tmpfile, err := utils.GetTestFile("cloud_test_file_*", []byte("hello there i see that you are a fan of bytes ?!@1")) // 50 bytes
	if err != nil {
		t.Fatal(err)
	}
	defer utils.GetTestFileCleanup(tmpfile)

	chunkSize := 10 // will give 5 chunks
	file, err := datastore.NewFile(tmpfile, tmpfile.Name(), chunkSize)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("File: %v", file)

	// Distribute file
	err = cloud.AddFile(file)
	if err != nil {
		t.Error(err)
	}
	t.Logf("Added File to network DataStore: %v.", cloud.Network().DataStore)

	t.Logf("Distributing file.")
	numReplicas := 1
	antiAffinity := true
	err = Distribute(file, cloud, numReplicas, antiAffinity)
	if err != nil {
		t.Error(err)
	}
	t.Logf("Final ChunkNodes: %v.", cloud.Network().ChunkNodes)
	t.Logf("Pretty ChunkNodes: %v.", cloud.ReadableChunkNodes())

	// TODO: proper comparison
	// chunks := file.Chunks.Chunks
	// chunkNodes := cloud.Network().ChunkNodes
	// if !(len(chunkNodes) == 2 && len(chunkNodes[chunks[0].ID]) == 2 && len(chunkNodes[chunks[1].ID]) == 2) {
	// 	t.Error("Unexpected ChunkNodes contents.")
	// }
}
