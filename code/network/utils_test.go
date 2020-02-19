package network

// testutils contains internal utility functions for tests.

import (
	"cloud/utils"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"time"
)

// GetTestFile returns a file with the given contents (caller must perform clean up on the file).
func GetTestFile(contents string) (*os.File, error) {
	tmpfile, err := ioutil.TempFile("", "cloud_test_file_*")
	if err != nil {
		return nil, err
	}

	path := tmpfile.Name()
	utils.GetLogger().Printf("Temporary filepath: %s", path)
	fileContentsBytes := []byte(contents)
	utils.GetLogger().Printf("Writing contents to temporary file: %s", contents)
	_, err = tmpfile.Write(fileContentsBytes)
	if err != nil {
		return nil, err
	}
	return tmpfile, nil
}

// CreateTestClouds makes a single cloud network but returns all the nodes' representations of the cloud.
// In a real life setting each cloud will run on a different machine.
// TODO: might want to test network representation (which should be the same), not the cloud representation.
// Also returns the storage directories.
// The caller must call os.RemoveAll(dir) to remove a directory.
func CreateTestClouds(numNodes int) ([]Cloud, []string, error) {
	tmpStorageDirs := make([]string, 0)
	for i := 0; i < numNodes; i++ {
		dir, err := ioutil.TempDir("", fmt.Sprintf("cloud_test_data_node_%d_", i))
		if err != nil {
			return nil, nil, err
		}
		tmpStorageDirs = append(tmpStorageDirs, dir)
	}

	key, err := GenerateKey()
	if err != nil {
		return nil, nil, err
	}
	nID, err := PublicKeyToID(&key.PublicKey)
	if err != nil {
		return nil, nil, err
	}

	// Create the first node that will begin the network.
	me := Node{
		ID:   nID,
		Name: "Node 1",
	}
	cloud := SetupNetwork(Network{
		Name: "my test network",
	}, me, key)
	cloud.SetConfig(CloudConfig{FileStorageDir: tmpStorageDirs[0]})
	clouds := []Cloud{cloud}
	cloud.ListenOnPort(0)
	go cloud.Accept()

	// Create the rest of the nodes.
	for i := 1; i < numNodes; i++ {
		snum := strconv.Itoa(i + 1)

		key, err := GenerateKey()
		if err != nil {
			return nil, nil, err
		}
		nID, err := PublicKeyToID(&key.PublicKey)
		if err != nil {
			return nil, nil, err
		}

		node := Node{
			ID:   nID,
			Name: "Node " + snum,
		}
		n, err := BootstrapToNetwork(cloud.MyNode().IP, node, key)
		if err != nil {
			return nil, nil, err
		}
		n.SetConfig(CloudConfig{FileStorageDir: tmpStorageDirs[i]})
		clouds = append(clouds, n)

		err = n.ListenOnPort(0)
		if err != nil {
			return nil, nil, err
		}
		go n.Accept()
	}
	time.Sleep(time.Millisecond * 100)
	return clouds, tmpStorageDirs, nil
}

// GenerateKey returns a random RSA private key.
func GenerateKey() (*rsa.PrivateKey, error) {
	pri, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	return pri, nil
}

// RemoveDirs removes all the directories in the list.
func RemoveDirs(dirs []string) {
	for _, dir := range dirs {
		os.RemoveAll(dir)
	}
}
