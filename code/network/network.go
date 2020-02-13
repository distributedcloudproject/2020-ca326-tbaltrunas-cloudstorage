package network

import (
	"cloud/comm"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"
)

type Node struct {
	ID string

	// Represented in ip:port format.
	// Example: 127.0.0.1:8081
	IP string

	// Display name of the node.
	Name string

	// Public key of the node.
	PublicKey crypto.PublicKey


	// client is the communication socket between us and the node.
	client comm.Client

	mutex sync.RWMutex
}

// Network is the general info of the network. Each node would have the same presentation of Network.
type Network struct {
	Name string
	Nodes []*Node

	// Require authentication for the network.
	RequireAuth bool

	// List of node IDs that are permitted to enter the Network. If `RequireAuth` is set, the IDs are enforced to be
	// generated based on the public key.
	AllowedIDs []string
}

// Cloud is the client's view of the Network. Contains client-specific information.
type Cloud struct {
	Network Network

	PendingNodes []*Node
	MyNode *Node
	PrivateKey *rsa.PrivateKey

	Listener net.Listener
	Mutex sync.RWMutex
	Port uint16

	SaveFunc func() io.Writer
}

type request struct {
	cloud *Cloud
	node *Node
}

func BootstrapToNetwork(ip string, me *Node, key *rsa.PrivateKey) (*Cloud, error) {
	// Establish connection with the target.
	client, err := comm.NewClientDial(ip, key)
	if err != nil {
		return nil, err
	}

	// Create a temporary node to represent the bootstrap node.
	node := &Node{
		IP: ip,
		client: client,
	}

	cloud := &Cloud{MyNode: me, PrivateKey: key}
	node.client.AddRequestHandler(createAuthRequestHandler(node, cloud))
	go node.client.HandleConnection()

	err = node.Authenticate(me)
	if err != nil {
		return nil, err
	}
	node.client.AddRequestHandler(createRequestHandler(node, cloud))

	// Update our info on the node.
	nodeInfo, err := node.NodeInfo()
	if err != nil {
		return nil, err
	}
	node.mutex.Lock()
	node.ID = nodeInfo.ID
	node.Name = nodeInfo.Name
	node.mutex.Unlock()

	network, err := node.NetworkInfo()
	if err != nil {
		return nil, err
	}

	node.client.Close()
	// Create dial connection to every node.
	for i := range network.Nodes {
		// It may not connect to some nodes (e.g offline) which is fine. That's why error is ignored.
		cloud.connectToNode(network.Nodes[i])
	}

	cloud.Network = network

	return cloud, nil
}

func SetupNetwork(me *Node, networkName string, key *rsa.PrivateKey) *Cloud {
	cloud := &Cloud{
		Network: Network{
			Name: "My new network",
			Nodes: []*Node{me},
		},
		MyNode: me,
		PrivateKey: key,
	}
	return cloud
}

func (n *Cloud) Listen(port int) error {
	var err error
	n.Listener, err = net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return err
	}
	return nil
}

func (n *Cloud) AcceptListener() {
	for {
		conn, err := n.Listener.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}

		client, err := comm.NewServerClient(conn, n.PrivateKey)
		if err != nil {
			fmt.Println(err)
			continue
		}
		node := &Node{
			IP: conn.RemoteAddr().String(),
			client: client,
		}
		node.client.AddRequestHandler(createAuthRequestHandler(node, n))
		n.PendingNodes = append(n.PendingNodes, node)
		go func(node *Node) {
			node.client.HandleConnection()

			n.Mutex.Lock()
			for _, c := range n.Network.Nodes {
				if c.ID == node.ID {
					c.client = nil
				}
			}
			n.Mutex.Unlock()
		}(node)
	}
}

func (n *Node) Ping() (string, error) {
	ping, err := n.client.SendMessage("ping", "ping")
	return ping[0].(string), err
}

func (r request) PingRequest(ping string) string {
	if ping == "ping" {
		return "pong"
	}
	return ""
}

func (n *Node) Online() bool {
	return n.client != nil
}

func PublicKeyToID(key *rsa.PublicKey) (string, error) {
	pub, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return "", err
	}
	sha := sha256.Sum256(pub)
	return hex.EncodeToString(sha[:]), nil
}