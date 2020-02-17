package distribution

import (
	"cloud/network"
	"cloud/datastore"
	"testing"
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

func createTestCloud(t *testing.T, numNodes int) *network.Cloud {
	genericFileStorageDir := filepath.Join("data", "node") // TODO: tempdir
	me := &network.Node{
		ID: "node1",
		Name: "Node 1",
		FileStorageDir: genericFileStorageDir + "1",
	}

	cloud := network.SetupNetwork(me, "My test network")
	cloud.Listen(0)
	go cloud.AcceptListener()
	me.IP = cloud.Listener.Addr().String()
	t.Logf("MyNode on cloud: %v.", cloud.MyNode)

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

		err = n.Listen(0)
		if err != nil {
			t.Error(err)
		}
		go n.AcceptListener()
	}
	time.Sleep(time.Millisecond * 100)

	t.Logf("Cloud with other nodes: %v.", cloud)
	t.Logf("Network: %v.", cloud.Network)
	for i := range cloud.Network.Nodes {
		t.Logf("Node %d: %v.", i, cloud.Network.Nodes[i])
	}
	return cloud
}

func TestFileDistribution(t *testing.T) {
	numNodes := 2
	cloud := createTestCloud(t, numNodes)
	t.Log(cloud)

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
	// TODO: check that all clouds have same datastore, with 1 file

	i := 0
	t.Logf("Saving chunk number: %d.", i)
	err = n.SaveChunk(file, i)
	if err != nil {
		t.Error(err)
	}
	t.Logf("Network with saved chunk: %v.", cloud.Network)
	t.Logf("Updated chunk-node locations: %v.", cloud.Network.FileChunkLocations)
}
