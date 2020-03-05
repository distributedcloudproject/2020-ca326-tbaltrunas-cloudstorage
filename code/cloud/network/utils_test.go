package network

import (
	"crypto/rand"
	"crypto/rsa"
	"strconv"
	"time"
)

// utils_test contains utility functions for network package tests.

// GenerateKey returns a random RSA private key.
func GenerateKey() (*rsa.PrivateKey, error) {
	pri, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	return pri, nil
}

// CreateTestClouds makes a single cloud network but returns all the nodes' representations of the cloud.
// In a real life setting each cloud will run on a different machine.
// TODO: might want to test network representation (which should be the same), not the cloud representation.
func CreateTestClouds(numNodes int) ([]Cloud, error) {
	key, err := GenerateKey()
	if err != nil {
		return nil, err
	}
	nID, err := PublicKeyToID(&key.PublicKey)
	if err != nil {
		return nil, err
	}

	// Create the first node that will begin the network.
	me := Node{
		ID:   nID,
		Name: "Node 1",
	}
	cloud := SetupNetwork(Network{
		Name: "my test network",
	}, me, key)
	clouds := []Cloud{cloud}
	cloud.ListenOnPort(0)
	go cloud.Accept()

	// Create the rest of the nodes.
	for i := 1; i < numNodes; i++ {
		snum := strconv.Itoa(i + 1)

		key, err := GenerateKey()
		if err != nil {
			return nil, err
		}
		nID, err := PublicKeyToID(&key.PublicKey)
		if err != nil {
			return nil, err
		}

		node := Node{
			ID:   nID,
			Name: "Node " + snum,
		}
		n, err := BootstrapToNetwork(cloud.MyNode().IP, node, key, CloudConfig{})
		if err != nil {
			return nil, err
		}
		clouds = append(clouds, n)

		err = n.ListenOnPort(0)
		if err != nil {
			return nil, err
		}
		go n.Accept()
	}
	time.Sleep(time.Millisecond * 100)
	return clouds, nil
}
