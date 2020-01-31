package network

import (
	"strconv"
	"testing"
	"time"
)

func TestNetworkPing(t *testing.T) {
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

	p, err := n2.Network.Nodes[0].Ping()
	if err != nil {
		t.Error(err)
	}
	if p != "pong" {
		t.Errorf("Ping() got %s; expected %s", p, "pong")
	}
	if len(cloud.Network.Nodes) != 2 {
		t.Errorf("network nodes: %v; expected %v", len(cloud.Network.Nodes), 1)
	}
	if len(n2.Network.Nodes) != 2 {
		t.Errorf("network nodes: %v; expected %v", len(cloud.Network.Nodes), 1)
	}
}

func TestNetworkBootstrap(t *testing.T) {
	me := &Node{
		ID: "1",
		Name: "test",
	}

	cloud := SetupNetwork(me, "My new network")
	cloud.Listen(0)
	go cloud.AcceptListener()
	me.IP = cloud.Listener.Addr().String()

	for i := 0; i < 4; i++ {
		_, err := BootstrapToNetwork(cloud.Listener.Addr().String(), &Node{
			ID: "id" + strconv.Itoa(i),
			Name: "Node " + strconv.Itoa(i+1),
		})
		if err != nil {
			t.Error(err)
		}
	}
	time.Sleep(time.Millisecond * 100)

	if len(cloud.Network.Nodes) != 5 {
		t.Errorf("network nodes: %v; expected %v", len(cloud.Network.Nodes), 5)
	}
}
