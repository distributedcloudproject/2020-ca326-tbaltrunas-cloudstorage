package network

import (
	"testing"
)

func TestNode_NetworkInfo(t *testing.T) {
	key, err := generateKey()
	if err != nil {
		t.Fatal(err)
	}
	nID, err := PublicKeyToID(&key.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	me := &Node{
		ID: nID,
		Name: "test",
	}

	cloud := SetupNetwork(me, "My new network", key)
	cloud.Listen(0)
	go cloud.AcceptListener()
	me.IP = cloud.Listener.Addr().String()

	key2, err := generateKey()
	if err != nil {
		t.Fatal(err)
	}
	nID2, err := PublicKeyToID(&key2.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	n2, err := BootstrapToNetwork(cloud.Listener.Addr().String(), &Node{
		ID: nID2,
		Name: "test2",
	}, key2)
	if err != nil {
		t.Error(err)
	}

	n, err := n2.Network.Nodes[0].NetworkInfo()
	if err != nil {
		t.Error(err)
	}
	if n.Name != "My new network" {
		t.Errorf("NetworkInfo() got %s; expected %s", n.Name, "My new network")
	}
}