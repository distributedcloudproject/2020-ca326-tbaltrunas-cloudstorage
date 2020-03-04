package network

import (
	"crypto/rsa"
	"testing"
)

func TestNode_NetworkInfo(t *testing.T) {
	key, nID, err := createKey()
	if err != nil {
		t.Fatal(err)
	}

	cloud := SetupNetwork(Network{
		Name:        "My new network",
		Whitelist:   false,
		RequireAuth: true,
	}, Node{Name: "test"}, key)

	cloud.ListenOnPort(0)
	go cloud.Accept()

	key2, _, err := createKey()
	if err != nil {
		t.Fatal(err)
	}
	n2, err := BootstrapToNetwork(cloud.MyNode().IP, Node{Name: "test2"}, key2, CloudConfig{})
	if err != nil {
		t.Fatal(err)
	}

	n, err := n2.GetCloudNode(nID).NetworkInfo()
	if err != nil {
		t.Error(err)
	}
	if n.Name != "My new network" {
		t.Errorf("NetworkInfo() got %s; expected %s", n.Name, "My new network")
	}
}

func createKey() (*rsa.PrivateKey, string, error) {
	key, err := generateKey()
	if err != nil {
		return nil, "", err
	}
	nID, err := PublicKeyToID(&key.PublicKey)
	if err != nil {
		return nil, "", err
	}
	return key, nID, nil
}
