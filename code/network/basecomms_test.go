package network

import (
	"testing"
)

func TestNode_NetworkInfo(t *testing.T) {
	network := &Network{
		Name: "My new network",
	}
	network.Listen(0)
	go network.AcceptListener()

	n2, err := BootstrapToNetwork(network.listener.Addr().String())
	if err != nil {
		t.Error(err)
	}

	n, err := n2.Nodes[0].NetworkInfo()
	if err != nil {
		t.Error(err)
	}
	if n.Name != "My new network" {
		t.Errorf("NetworkInfo() got %s; expected %s", n.Name, "My new network")
	}
}
