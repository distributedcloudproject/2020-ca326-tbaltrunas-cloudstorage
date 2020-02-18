package network

import (
	"cloud/datastore"
	"cloud/testutils"
	"testing"
	"fmt"
	"reflect"
	"os"
	"io/ioutil"
	"strconv"
	"time"
)

// Returns a file with contents and a reader of that file (caller must perform clean up on it.)
func getTestFile(t *testing.T) (*datastore.File, *os.File) {
	tmpfile, err := ioutil.TempFile("", "cloud_test_file_*")
	if err != nil {
		t.Error(err)
	}
	
	path := tmpfile.Name()
	t.Logf("Temporary filepath: %s", path)
	fileContents := "hellothere i see you are a fan of bytes?"  // 40 bytes
	fileContentsBytes := []byte(fileContents)
	t.Logf("Writing contents to temporary file: %s", fileContents)
	_, err = tmpfile.Write(fileContentsBytes)
	if err != nil {
		t.Error(err)
	}

	chunkSize := 10  // will give 4 chunks
	file, err := datastore.NewFile(tmpfile, path, chunkSize)
	if err != nil {
		t.Error(err)
	}
	return file, tmpfile
}

// createTestClouds makes a single cloud network but returns all the nodes' representations of the cloud.
// In a real life setting each cloud will run on a different machine.
// TODO: might want to test network representation (which should be the same), not the cloud representation.
// Also returns the storage directories.
// The caller must call os.RemoveAll(dir) to remove a directory.
func createTestClouds(t *testing.T, numNodes int) ([]*Cloud, []string) {
	tmpStorageDirs := make([]string, 0)
	for i := 0; i < numNodes; i++ {
		dir, err := ioutil.TempDir("", fmt.Sprintf("cloud_test_data_node_%d_", i))
		if err != nil {
			t.Fatal(err)
		}
		tmpStorageDirs = append(tmpStorageDirs, dir)
	}

	key, err := testutils.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	nID, err := PublicKeyToID(&key.PublicKey)
	if err != nil {
		t.Fatal(err)
	}

	// Create the first node that will begin the network.
	me := &Node{
		ID: nID,
		Name: "Node 1",
		FileStorageDir: tmpStorageDirs[0],
	}
	cloud := SetupNetwork(me, "My test network", key)
	clouds := []*Cloud{cloud}
	cloud.Listen(0)
	go cloud.AcceptListener()
	me.IP = cloud.Listener.Addr().String()

	// Create the rest of the nodes.
	for i := 1; i < numNodes; i++ {
		t.Log(i)
		snum := strconv.Itoa(i+1)

		key, err := testutils.GenerateKey()
		if err != nil {
			t.Fatal(err)
		}
		nID, err := PublicKeyToID(&key.PublicKey)
		if err != nil {
			t.Fatal(err)
		}

		n, err := BootstrapToNetwork(cloud.Listener.Addr().String(), &Node{
			ID: nID,
			Name: "Node " + snum,
			FileStorageDir: tmpStorageDirs[i],
		}, key)
		if err != nil {
			t.Error(err)
		}
		clouds = append(clouds, n)

		err = n.Listen(0)
		if err != nil {
			t.Error(err)
		}
		go n.AcceptListener()
	}
	time.Sleep(time.Millisecond * 100)
	return clouds, tmpStorageDirs
}

func TestNode_AddFileSaveChunk(t *testing.T) {
	numNodes := 4
	clouds, tmpStorageDirs := createTestClouds(t, numNodes)
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

	file, f := getTestFile(t)
	defer os.Remove(f.Name())
	defer f.Close()
	t.Logf("File: %v", file)

	n := cloud.Network.Nodes[0]
	t.Logf("Node: %v.", n)

	err := n.AddFile(file)
	if err != nil {
		t.Error(err)
	}
	t.Logf("Network with added file: %v.", cloud.Network)
	t.Logf("Updated datastore: %v.", cloud.Network.DataStore)
	// Check that we have a required DataStore
	if !(len(cloud.Network.DataStore.Files) == 1 && cloud.Network.DataStore.Files[0].ID == file.ID) {
		 t.Error("DataStore does not have the expected contents (the required file).")
	}

	// Check that all clouds have same DataStore
	ds := cloud.Network.DataStore
	for _, c := range clouds {
		dsOther := c.Network.DataStore
		t.Logf("DataStore in another cloud representation: %v.", dsOther)
		if len(dsOther.Files) == 0 {
			t.Error("DataStore is empty.")
		}
		t.Logf("File in another cloud representation: %v.", dsOther.Files[0])
		if ds.Files[0].ID != dsOther.Files[0].ID {
			t.Error("DataStores not matching across cloud representations.")
		}
	}

	cn := cloud.Network.Nodes
	t.Log("Distributing chunks.")
	// TODO: move to a function on its own
	for i := 0; i < file.Chunks.NumChunks; i++ {
		t.Logf("Distributing chunk: %d (ID: %v), on node: %v", i, file.Chunks.Chunks[i].ID, cn[i])
		err = cn[i].SaveChunk(file, i)
		if err != nil {
			t.Error(err)
		}
	}

	t.Logf("Network with saved chunk: %v.", cloud.Network)
	t.Logf("Updated chunk-node locations: %v.", cloud.Network.ChunkNodes)
	// Check that we have a required ChunkNodes.
	chunks := file.Chunks.Chunks
	expectedChunkNodes := ChunkNodes{
		chunks[0].ID: []string{cn[0].ID},
		chunks[1].ID: []string{cn[1].ID},
		chunks[2].ID: []string{cn[2].ID},
		chunks[3].ID: []string{cn[3].ID},
	}
	t.Logf("Expected ChunkNodes: %v.", expectedChunkNodes)
	// Note that DeepEqual has arguments against using it.
	// https://stackoverflow.com/a/45222521
	// An alternative struct comparison method may be needed in the future.
	if !reflect.DeepEqual(cloud.Network.ChunkNodes, expectedChunkNodes) {
		t.Error("ChunkNodes does not have the expected contents.")
	}

	// Check that all clouds have same ChunkNodes
	chunkLocations := cloud.Network.ChunkNodes
	for _, c := range clouds {
		chunkLocationsOther := c.Network.ChunkNodes
		t.Logf("ChunkNodes in another cloud representation: %v.", chunkLocationsOther)
		if !reflect.DeepEqual(chunkLocations, chunkLocationsOther) {
			t.Error("ChunkNodes not matching across cloud representations.")
		}
	}
}
