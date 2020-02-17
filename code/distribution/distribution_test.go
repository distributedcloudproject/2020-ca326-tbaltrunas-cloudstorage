package distribution

import (
	"cloud/network"
	"cloud/datastore"
	"testing"
	"reflect"
	"os"
	"io/ioutil"
	"strconv"
	"time"
	"path/filepath"
)

// Returns a file with contents and a reader of that file (caller must perform clean up on it.)
func getTestFile(t *testing.T) (*datastore.File, *os.File) {
	tmpfile, err := ioutil.TempFile("", "cloud_test_file_*")
	if err != nil {
		t.Error(err)
	}
	
	path := tmpfile.Name()
	t.Logf("Temporary filepath: %s", path)
	fileContents := "hellothere"  // 10 bytes
	fileContentsBytes := []byte(fileContents)
	t.Logf("Writing contents to temporary file: %s", fileContents)
	_, err = tmpfile.Write(fileContentsBytes)
	if err != nil {
		t.Error(err)
	}

	chunkSize := 5
	file, err := datastore.NewFile(tmpfile, path, chunkSize)
	if err != nil {
		t.Error(err)
	}
	return file, tmpfile
}

// createTestClouds makes a single cloud network but returns all the nodes' representations of the cloud.
// TODO: might want to test network representation (which should be the same), not the cloud representation.
func createTestClouds(t *testing.T, numNodes int) []*network.Cloud {
	genericFileStorageDir := filepath.Join("data", "node") // TODO: tempdir
	me := &network.Node{
		ID: "node1",
		Name: "Node 1",
		FileStorageDir: genericFileStorageDir + "1",
	}

	cloud := network.SetupNetwork(me, "My test network")
	clouds := []*network.Cloud{cloud}
	cloud.Listen(0)
	go cloud.AcceptListener()
	me.IP = cloud.Listener.Addr().String()

	for i := 1; i < numNodes; i++ {
		t.Log(i)
		snum := strconv.Itoa(i+1)
		n, err := network.BootstrapToNetwork(cloud.Listener.Addr().String(), &network.Node{
			ID: "node" + snum,
			Name: "Node " + snum,
			FileStorageDir: genericFileStorageDir + snum,
		})
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
	return clouds
}

func TestFileDistribution(t *testing.T) {
	numNodes := 3
	clouds := createTestClouds(t, numNodes)
	t.Logf("Test clouds: %v.", clouds)
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

	// i := 0
	// t.Logf("Saving chunk number: %d.", i)
	// err = n.SaveChunk(file, i)
	// if err != nil {
	// 	t.Error(err)
	// }
	cn := cloud.Network.Nodes
	t.Log("Distributing chunks.")
	for i := 0; i < file.Chunks.NumChunks; i++ {
		n := cn[i % len(cn)]
		t.Logf("Distributing chunk: %d, on node: %v", i, n)
		err = n.SaveChunk(file, i)
		if err != nil {
			t.Error(err)
		}
	}

	t.Logf("Network with saved chunk: %v.", cloud.Network)
	t.Logf("Updated chunk-node locations: %v.", cloud.Network.FileChunkLocations)
	// Check that all clouds have same FileChunkLocations
	chunkLocations := cloud.Network.FileChunkLocations
	for _, c := range clouds {
		chunkLocationsOther := c.Network.FileChunkLocations
		t.Logf("FileChunkLocations in another cloud representation: %v.", chunkLocationsOther)
		// TODO: datastore comparison method
		if !reflect.DeepEqual(chunkLocations, chunkLocationsOther) {
			t.Error("FileChunkLocations not matching across cloud representations.")
		}
	}
}
