package network

import (
	"cloud/comm"
	"fmt"
	"net"
	"strconv"
	"sync"
)

type Node struct {
	// Represented in ip:port format.
	// Example: 127.0.0.1:8081
	IP string

	// Display name of the node.
	Name string

	Authenticated bool


	// client is the communication socket between us and the node.
	client comm.Client

	mutex sync.RWMutex
}

type Network struct {
	Name string

	Nodes []*Node

	listener net.Listener

	mutex sync.RWMutex
}

type request struct {
	network *Network
	node *Node
}

func BootstrapToNetwork(ip string) (*Network, error) {
	client, err := comm.NewClientDial(ip)
	if err != nil {
		return nil, err
	}

	node := &Node{
		IP: ip,
		Authenticated: true,
		client: client,
	}

	network := &Network{Nodes: []*Node{node}}
	network.Nodes[0].client.AddRequestHandler(createAuthRequestHandler(node, network))
	go network.Nodes[0].client.HandleConnection()

	err = node.Authenticate()
	if err != nil {
		return nil, err
	}
	node.client.AddRequestHandler(createRequestHandler(node, network))

	return network, nil
}

func createRequestHandler(node *Node, network *Network) func(string) interface{} {
	r := request{
		network: network,
		node: node,
	}

	return func(message string) interface{} {
		switch message {
		case "ping": return r.PingRequest
		case NetworkInfoMsg: return r.OnNetworkInfoRequest
		}
		return nil
	}
}

func (n *Network) Listen(port int) error {
	var err error
	n.listener, err = net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return err
	}
	return nil
}

func (n *Network) AcceptListener() {
	for {
		conn, err := n.listener.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}

		node := &Node{
			IP: conn.RemoteAddr().String(),
			client: comm.NewClient(conn),
		}
		node.client.AddRequestHandler(createAuthRequestHandler(node, n))
		n.Nodes = append(n.Nodes, node)
		go node.client.HandleConnection()
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
