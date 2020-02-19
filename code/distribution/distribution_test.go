package distribution

import (
	"cloud/testutils"
	"cloud/datastore"
	"testing"
	"os"
)

func TestChunkDistribution(t *testing.T) {
	numNodes := 2
	clouds, tmpStorageDirs, err := testutils.CreateTestClouds(numNodes)
	if err != nil {
		t.Fatal(err)
	}
	defer testutils.RemoveDirs(tmpStorageDirs)

	t.Logf("Test clouds: %v.", clouds)
	t.Logf("Storage locations for clouds: %v.", tmpStorageDirs)

	cloud := clouds[0]
	t.Logf("Main cloud: %v.", cloud)
	t.Logf("MyNode on cloud: %v.", cloud.MyNode)
	t.Logf("Cloud with other nodes: %v.", cloud)
	t.Logf("Network: %v.", cloud.Network)
	for i := range cloud.Network.Nodes {
		t.Logf("Node %d: %v.", i, cloud.Network.Nodes[i])
	}

	tmpfile, err := testutils.GetTestFile("hellothere") // 10 bytes
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	defer tmpfile.Close()
	chunkSize := 5  // will give 2 chunks
	file, err := datastore.NewFile(tmpfile, tmpfile.Name(), chunkSize)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("File: %v", file)
	t.Logf("File: %v", file)

	n := cloud.Network.Nodes[0]
	t.Logf("Node: %v.", n)

	// Distribute file
	err = cloud.MyNode.AddFile(file)
	if err != nil {
		t.Error(err)
	}
	t.Logf("Added File to network DataStore: %v.", cloud.Network.DataStore)

	t.Logf("Distributing file.")
	err = Distribute(file, cloud)
	if err != nil {
		t.Error(err)
	}
	t.Logf("Final ChunkNodes: %v.", cloud.Network.ChunkNodes)

	// TODO: proper comparison
	chunks := file.Chunks.Chunks
	chunkNodes := cloud.Network.ChunkNodes
	if !(len(chunkNodes) == 2 && len(chunkNodes[chunks[0].ID]) == 2 && len(chunkNodes[chunks[1].ID]) == 2) {
		t.Error("Unexpected ChunkNodes contents.")
	}
}