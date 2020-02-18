package distribution

import (
	"cloud/network"
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
func createTestClouds(t *testing.T, numNodes int) ([]*network.Cloud, []string) {
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
	nID, err := network.PublicKeyToID(&key.PublicKey)
	if err != nil {
		t.Fatal(err)
	}

	// Create the first node that will begin the network.
	me := &network.Node{
		ID: nID,
		Name: "Node 1",
		FileStorageDir: tmpStorageDirs[0],
	}
	cloud := network.SetupNetwork(me, "My test network", key)
	clouds := []*network.Cloud{cloud}
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
		nID, err := network.PublicKeyToID(&key.PublicKey)
		if err != nil {
			t.Fatal(err)
		}

		n, err := network.BootstrapToNetwork(cloud.Listener.Addr().String(), &network.Node{
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

func removeDirs(dirs []string) {
	for _, dir := range dirs {
		os.RemoveAll(dir)
	}
}

func TestFileDistribution(t *testing.T) {
	numNodes := 4
	clouds, tmpStorageDirs := createTestClouds(t, numNodes)
	defer removeDirs(tmpStorageDirs)

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
	if !(len(cloud.Network.DataStore.Files) == 1 && 
		 reflect.DeepEqual(cloud.Network.DataStore.Files[0].Chunks, file.Chunks)) {
		 t.Error("DataStore does not have the expected contents (the required file).")
	}

	// Check that all clouds have same DataStore
	ds := cloud.Network.DataStore
	for _, c := range clouds {
		dsOther := c.Network.DataStore
		t.Logf("DataStore in another cloud representation: %v.", dsOther)
		// TODO: datastore comparison method
		if !reflect.DeepEqual(ds.Files[0].Chunks, dsOther.Files[0].Chunks) {
			t.Error("DataStores not matching across cloud representations.")
		}
	}

	cn := cloud.Network.Nodes
	t.Log("Distributing chunks.")
	// TODO: move to a function on its own
	for i := 0; i < file.Chunks.NumChunks; i++ {
		n := cn[i]
		t.Logf("Distributing chunk: %d (ID: %v), on node: %v", i, file.Chunks.Chunks[i].ID, n)
		err = n.SaveChunk(file, i)
		if err != nil {
			t.Error(err)
		}
	}

	t.Logf("Network with saved chunk: %v.", cloud.Network)
	t.Logf("Updated chunk-node locations: %v.", cloud.Network.FileChunkLocations)
	// Check that we have a required FileChunkLocations.
	chunks := file.Chunks.Chunks
	expectedFileChunkLocations := network.FileChunkLocations{
		chunks[0].ID: []string{cn[0].ID},
		chunks[1].ID: []string{cn[1].ID},
		chunks[2].ID: []string{cn[2].ID},
		chunks[3].ID: []string{cn[3].ID},
	}
	t.Logf("Expected FileChunkLocations: %v.", expectedFileChunkLocations)
	if !reflect.DeepEqual(cloud.Network.FileChunkLocations, expectedFileChunkLocations) {
		t.Error("FileChunkLocations does not have the expected contents.")
	}

	// Check that all clouds have same FileChunkLocations
	chunkLocations := cloud.Network.FileChunkLocations
	for _, c := range clouds {
		chunkLocationsOther := c.Network.FileChunkLocations
		t.Logf("FileChunkLocations in another cloud representation: %v.", chunkLocationsOther)
		if !reflect.DeepEqual(chunkLocations, chunkLocationsOther) {
			t.Error("FileChunkLocations not matching across cloud representations.")
		}
	}
}
