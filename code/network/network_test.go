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

	for i := 0; i < 4; i++ {
		key2, err := generateKey()
		if err != nil {
			t.Fatal(err)
		}
		nID2, err := PublicKeyToID(&key2.PublicKey)
		if err != nil {
			t.Fatal(err)
		}
		n, err := BootstrapToNetwork(cloud.Listener.Addr().String(), &Node{
			ID: nID2,
			Name: "Node " + strconv.Itoa(i+1),
		}, key2)
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
	var clouds []*Cloud

	for i := 0; i < 4; i++ {
		listener, err := net.Listen("tcp", ":0")
		if err != nil {
			t.Error(err)
		}

		key2, err := generateKey()
		if err != nil {
			t.Fatal(err)
		}

		nID2, err := PublicKeyToID(&key2.PublicKey)
		if err != nil {
			t.Fatal(err)
		}

		n, err := BootstrapToNetwork(cloud.Listener.Addr().String(), &Node{
			ID: nID2,
			Name: "Node " + strconv.Itoa(i+1),
			IP: listener.Addr().String(),
		}, key2)
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

func TestNetworkWhitelist(t *testing.T) {
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
	cloud.Network.Whitelist = true

	cloud.Listen(0)
	go cloud.AcceptListener()
	me.IP = cloud.Listener.Addr().String()

	// Node 2, should not be permitted into the network.
	key2, err := generateKey()
	if err != nil {
		t.Fatal(err)
	}
	nID2, err := PublicKeyToID(&key2.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	_, err = BootstrapToNetwork(cloud.Listener.Addr().String(), &Node{
		ID: nID2,
		Name: "test2",
	}, key2)
	if err == nil {
		t.Error("Node2 connected to whitelisted network.")
	}

	if len(cloud.Network.Nodes) != 1 {
		t.Errorf("network nodes: %v; expected %v", len(cloud.Network.Nodes), 1)
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
	_, err = BootstrapToNetwork(cloud.Listener.Addr().String(), &Node{
		ID: nID3,
		Name: "test3",
	}, key3)
	if err != nil {
		t.Error("Node3 failed to connect to network: ", err)
	}

	if len(cloud.Network.Nodes) != 2 {
		t.Errorf("network nodes: %v; expected %v", len(cloud.Network.Nodes), 2)
	}
}

func generateKey() (*rsa.PrivateKey, error) {
	pri, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	return pri, nil
}