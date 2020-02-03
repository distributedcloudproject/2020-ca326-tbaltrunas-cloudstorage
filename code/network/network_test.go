package network

import (
	"net"
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
		n, err := BootstrapToNetwork(cloud.Listener.Addr().String(), &Node{
			ID: "id" + strconv.Itoa(i),
			Name: "Node " + strconv.Itoa(i+1),
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

	if len(cloud.Network.Nodes) != 5 {
		t.Errorf("network nodes: %v; expected %v", len(cloud.Network.Nodes), 5)
	}
}

func TestNetworkAddNode(t *testing.T) {
	me := &Node{
		ID: "1",
		Name: "test",
	}

	cloud := SetupNetwork(me, "My new network")
	cloud.Listen(0)
	go cloud.AcceptListener()
	me.IP = cloud.Listener.Addr().String()
	var clouds []*Cloud

	for i := 0; i < 4; i++ {
		listener, err := net.Listen("tcp", ":0")
		if err != nil {
			t.Error(err)
		}

		n, err := BootstrapToNetwork(cloud.Listener.Addr().String(), &Node{
			ID: "id" + strconv.Itoa(i),
			Name: "Node " + strconv.Itoa(i+1),
			IP: listener.Addr().String(),
		})
		if err != nil {
			t.Error(err)
		}

		n.Listener = listener
		go n.AcceptListener()
		clouds = append(clouds, n)
	}
	time.Sleep(time.Millisecond * 100)

	for _, c := range clouds {
		if len(c.Network.Nodes) != 5 {
			t.Errorf("network nodes: %v; expected %v", len(cloud.Network.Nodes), 5)
		}
	}
}