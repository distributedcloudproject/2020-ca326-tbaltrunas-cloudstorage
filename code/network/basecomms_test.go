package network

import (
	"testing"
)

func TestNode_NetworkInfo(t *testing.T) {
	me := &Node{
		ID: "1",
		Name: "test",
	}

	cloud := SetupNetwork(me, "My new network")
	cloud.Listen(0)
	go cloud.AcceptListener()
	me.IP = cloud.Listener.Addr().String()

	n2, err := BootstrapToNetwork(cloud.Listener.Addr().String(), &Node{
		ID: "2",
		Name: "test2",
	})
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