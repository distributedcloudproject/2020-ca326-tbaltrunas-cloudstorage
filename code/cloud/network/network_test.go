package network

import (
	"crypto/rand"
	"crypto/rsa"
	"net"
	"strconv"
	"testing"
	"time"
)

func TestNetworkPing(t *testing.T) {
	key, err := generateKey()
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

	key2, err := generateKey()
	if err != nil {
		t.Fatal(err)
	}
	n2, err := BootstrapToNetwork(cloud.MyNode().IP, Node{Name: "test2"}, key2, CloudConfig{})
	if err != nil {
		t.Fatal(err)
	}

	p, err := n2.GetCloudNode(cloud.MyNode().ID).Ping()
	if err != nil {
		t.Error(err)
	}
	if p != "pong" {
		t.Errorf("Ping() got %s; expected %s", p, "pong")
	}
	if onlineNodes := cloud.OnlineNodesNum(); onlineNodes != 2 {
		t.Errorf("network nodes: %v; expected %v", onlineNodes, 2)
	}
	if onlineNodes := n2.OnlineNodesNum(); onlineNodes != 2 {
		t.Errorf("network nodes: %v; expected %v", onlineNodes, 2)
	}
}

func TestNetworkBootstrap(t *testing.T) {
	key, _, err := createKey()
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

	for i := 0; i < 4; i++ {
		key2, err := generateKey()
		if err != nil {
			t.Fatal(err)
		}

		n2, err := BootstrapToNetwork(cloud.MyNode().IP, Node{Name: "Node " + strconv.Itoa(i+1)}, key2, CloudConfig{})
		if err != nil {
			t.Fatal(err)
		}

		err = n2.ListenOnPort(0)
		if err != nil {
			t.Error(err)
		}
		go n2.Accept()
	}
	time.Sleep(time.Millisecond * 100)

	if onlineNodes := cloud.OnlineNodesNum(); onlineNodes != 5 {
		t.Errorf("network nodes: %v; expected %v", onlineNodes, 5)
	}
}

func TestNetworkAddNode(t *testing.T) {
	key, _, err := createKey()
	if err != nil {
		t.Fatal(err)
	}

	cloud := SetupNetwork(Network{
		Name:        "My new network",
		Whitelist:   true,
		RequireAuth: true,
	}, Node{Name: "test"}, key)
	cloud.ListenOnPort(0)
	go cloud.Accept()
	var clouds []Cloud

	for i := 0; i < 4; i++ {
		listener, err := net.Listen("tcp", ":0")
		if err != nil {
			t.Error(err)
		}

		key2, err := generateKey()
		if err != nil {
			t.Fatal(err)
		}
		id2, _ := PublicKeyToID(&key2.PublicKey)
		cloud.AddToWhitelist(id2)

		n, err := BootstrapToNetwork(cloud.MyNode().IP, Node{
			Name: "Node " + strconv.Itoa(i+1),
			IP:   listener.Addr().String(),
		}, key2, CloudConfig{})
		if err != nil {
			t.Error(err)
		}

		go n.AcceptUsingListener(listener)
		clouds = append(clouds, n)
	}
	time.Sleep(time.Millisecond * 100)

	for _, c := range clouds {
		if nodes := c.NodesNum(); nodes != 5 {
			t.Errorf("network(%v) nodes: %v; expected %v", c.MyNode().ID, nodes, 5)
		}
	}
}

func TestNetworkWhitelist(t *testing.T) {
	key, _, err := createKey()
	if err != nil {
		t.Fatal(err)
	}

	cloud := SetupNetwork(Network{
		Name:        "My new network",
		Whitelist:   true,
		RequireAuth: true,
	}, Node{Name: "test"}, key)

	cloud.ListenOnPort(0)
	go cloud.Accept()

	// Node 2, should not be permitted into the network.
	key2, err := generateKey()
	if err != nil {
		t.Fatal(err)
	}
	_, err = BootstrapToNetwork(cloud.MyNode().IP, Node{Name: "test2"}, key2, CloudConfig{})
	if err == nil {
		t.Error("Node2 connected to whitelisted network.")
	}

	if nodes := cloud.NodesNum(); nodes != 1 {
		t.Errorf("network nodes: %v; expected %v", nodes, 1)
	}

	// Node 3, should  be permitted into the network.
	key3, err := generateKey()
	if err != nil {
		t.Fatal(err)
	}
	nID3, err := PublicKeyToID(&key3.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	cloud.AddToWhitelist(nID3)
	_, err = BootstrapToNetwork(cloud.MyNode().IP, Node{Name: "test3"}, key3, CloudConfig{})
	if err != nil {
		t.Error("Node3 failed to connect to network: ", err)
	}

	if nodes := cloud.NodesNum(); nodes != 2 {
		t.Errorf("network nodes: %v; expected %v", nodes, 2)
	}
}

func TestNetworkLocalNode(t *testing.T) {
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

	msg, err := cloud.GetCloudNode(nID).Ping()
	if err != nil {
		t.Fatal("Failed to ping:", err)
	}
	if msg != "pong" {
		t.Fatal("me.Ping() want pong; got", msg)
	}
}

func generateKey() (*rsa.PrivateKey, error) {
	pri, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	return pri, nil
}
