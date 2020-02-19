package testutils

// testutils contains internal utility functions for tests.

import (
	"cloud/network"
	"cloud/utils"
	"os"
	"fmt"
	"strconv"
	"time"
	"io/ioutil"
	"crypto/rand"
	"crypto/rsa"
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
func CreateTestClouds(numNodes int) ([]*network.Cloud, []string, error) {
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
	nID, err := network.PublicKeyToID(&key.PublicKey)
	if err != nil {
		return nil, nil, err
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
		snum := strconv.Itoa(i+1)

		key, err := GenerateKey()
		if err != nil {
			return nil, nil, err
		}
		nID, err := network.PublicKeyToID(&key.PublicKey)
		if err != nil {
			return nil, nil, err
		}

		node := &network.Node{
			ID: nID,
			Name: "Node " + snum,
			FileStorageDir: tmpStorageDirs[i],
		}
		n, err := network.BootstrapToNetwork(cloud.Listener.Addr().String(), node, key)
		if err != nil {
			return nil, nil, err
		}
		clouds = append(clouds, n)

		err = n.Listen(0)
		if err != nil {
			return nil, nil, err
		}
		go n.AcceptListener()
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
