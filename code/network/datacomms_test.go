package network

import (
	"cloud/datastore"
	"os"
	"reflect"
	"testing"
)

func TestNode_AddFileSaveChunk(t *testing.T) {
	numNodes := 4
	clouds, tmpStorageDirs, err := CreateTestClouds(numNodes)
	if err != nil {
		t.Fatal(err)
	}
	defer RemoveDirs(tmpStorageDirs)

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

	tmpfile, err := GetTestFile("hellothere i see you are a fan of bytes?") // 40 bytes
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	defer tmpfile.Close()
	chunkSize := 10 // will give 4 chunks
	file, err := datastore.NewFile(tmpfile, tmpfile.Name(), chunkSize)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("File: %v", file)

	n := nodes[0]
	t.Logf("Node: %v.", n)

	err = cloud.AddFile(file)
	if err != nil {
		t.Error(err)
	}
	t.Logf("Network with added file: %v.", cloud.Network())
	t.Logf("Updated datastore: %v.", cloud.Network().DataStore)
	// Check that we have a required DataStore
	if !(len(cloud.Network().DataStore.Files) == 1 && cloud.Network().DataStore.Files[0].ID == file.ID) {
		t.Error("DataStore does not have the expected contents (the required file).")
	}

	// Check that all clouds have same DataStore
	ds := cloud.Network().DataStore
	for _, c := range clouds {
		dsOther := c.Network().DataStore
		t.Logf("DataStore in another cloud representation: %v.", dsOther)
		if len(dsOther.Files) == 0 {
			t.Error("DataStore is empty.")
			continue
		}
		t.Logf("File in another cloud representation: %v.", dsOther.Files[0])
		if ds.Files[0].ID != dsOther.Files[0].ID {
			t.Error("DataStores not matching across cloud representations.")
		}
	}

	t.Log("Distributing chunks.")
	// TODO: move to a function on its own
	for i := 0; i < file.Chunks.NumChunks; i++ {
		t.Logf("Distributing chunk: %d (ID: %v), on node: %v", i, file.Chunks.Chunks[i].ID, cloud.Network().Nodes[i])
		err = cloud.GetCloudNode(cloud.Network().Nodes[i].ID).SaveChunk(file, i)
		if err != nil {
			t.Error(err)
		}
	}

	// Check that we have a required ChunkNodes.
	t.Logf("Updated chunk-node locations: %v.", cloud.Network().ChunkNodes)
	chunks := file.Chunks.Chunks
	actualChunkNodes := cloud.Network().ChunkNodes
	expectedChunkNodes := ChunkNodes{
		chunks[0].ID: []string{cloud.Network().Nodes[0].ID},
		chunks[1].ID: []string{cloud.Network().Nodes[1].ID},
		chunks[2].ID: []string{cloud.Network().Nodes[2].ID},
		chunks[3].ID: []string{cloud.Network().Nodes[3].ID},
	}
	t.Logf("Expected ChunkNodes: %v.", expectedChunkNodes)
	// Note that DeepEqual has arguments against using it.
	// https://stackoverflow.com/a/45222521
	// An alternative struct comparison method may be needed in the future.
	// if !reflect.DeepEqual(cloud.Network.ChunkNodes, expectedChunkNodes) {
	// 	t.Error("ChunkNodes does not have the expected contents.")
	// }
	// FIXME: DeepEqual returns false. Need a better method.
	// Quick fix down below:
	if len(actualChunkNodes) != len(expectedChunkNodes) {
		t.Error("Actual and expected ChunkNodes do not match.")
	}
	for k := range expectedChunkNodes {
		v := expectedChunkNodes[k]
		va := actualChunkNodes[k]
		for i, _ := range v {
			if v[i] != va[i] {
				t.Errorf("Element mismatch: %v, %v.", v[i], va[i])
			}
		}
	}
	for k := range actualChunkNodes {
		v := actualChunkNodes[k]
		va := expectedChunkNodes[k]
		for i, _ := range v {
			if v[i] != va[i] {
				t.Errorf("Element mismatch: %v, %v.", v[i], va[i])
			}
		}
	}

	// Check that all clouds have same ChunkNodes
	chunkLocations := cloud.Network().ChunkNodes
	for _, c := range clouds {
		chunkLocationsOther := c.Network().ChunkNodes
		t.Logf("ChunkNodes in another cloud representation: %v.", chunkLocationsOther)
		if !reflect.DeepEqual(chunkLocations, chunkLocationsOther) {
			t.Error("ChunkNodes not matching across cloud representations.")
		}
	}
}
