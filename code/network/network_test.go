package network

import (
	"testing"
	"time"
)

func TestNetworkPing(t *testing.T) {
	network := &Network{
		Name: "Test",
	}
	network.Listen(0)
	go network.AcceptListener()

	n2, err := BootstrapToNetwork(network.listener.Addr().String())
	if err != nil {
		t.Error(err)
	}

	p, err := n2.Nodes[0].Ping()
	if err != nil {
		t.Error(err)
	}
	if p != "pong" {
		t.Errorf("Ping() got %s; expected %s", p, "pong")
	}

	if len(network.Nodes) != 1 {
		t.Errorf("network nodes: %v; expected %v", len(network.Nodes), 1)
	}
}

func TestNetworkBootstrap(t *testing.T) {
	network := &Network{
		Name: "Test",
	}
	network.Listen(0)
	go network.AcceptListener()

	for i := 0; i < 100; i++ {
		_, err := BootstrapToNetwork(network.listener.Addr().String())
		if err != nil {
			t.Error(err)
		}
	}
	time.Sleep(time.Millisecond * 100)

	if len(network.Nodes) != 100 {
		t.Errorf("network nodes: %v; expected %v", len(network.Nodes), 100)
	}
}
