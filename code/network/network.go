package network

import (
	"cloud/comm"
	"cloud/utils"
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


	// client is the communication socket between us and the node.
	client comm.Client

	mutex sync.RWMutex
}

// Network is the general info of the network. Each node would have the same presentation of Network.
type Network struct {
	Name string
	Nodes []*Node
}

// Cloud is the client's view of the Network. Contains client-specific information.
type Cloud struct {
	Network Network

	PendingNodes []*Node
	MyNode *Node

	Listener net.Listener
	Mutex sync.RWMutex
	Port uint16

	SaveFunc func() io.Writer
}

type request struct {
	cloud *Cloud
	node *Node
}

func BootstrapToNetwork(ip string, me *Node) (*Cloud, error) {
	// Establish connection with the target.
	utils.GetLogger().Printf("[INFO] Bootstrapping with ip: %v, and node: %v.", ip, me)
	client, err := comm.NewClientDial(ip)
	if err != nil {
		return nil, err
	}

	// Create a temporary node to represent the bootstrap node.
	node := &Node{
		IP: ip,
		client: client,
	}
	utils.GetLogger().Printf("[DEBUG] Remote node: %v.", node)

	cloud := &Cloud{MyNode: me}
	utils.GetLogger().Printf("[DEBUG] Initial cloud: %v.", cloud)
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
	utils.GetLogger().Printf("[INFO] Updated remote node info: %v.", node)

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
	utils.GetLogger().Printf("[INFO] Cloud with new network: %v.", cloud)

	return cloud, nil
}

func SetupNetwork(me *Node, networkName string) *Cloud {
	utils.GetLogger().Printf("[INFO] Setting up network with name: %v, and initial node: %v.", networkName, me)
	cloud := &Cloud{
		Network: Network{
			Name: "My new network",
			Nodes: []*Node{me},
		},
		MyNode: me,
	}
	me.client = comm.NewLocalClient()
	me.client.AddRequestHandler(createRequestHandler(me, cloud))
	return cloud
}

func (n *Cloud) Listen(port int) error {
	utils.GetLogger().Printf("[INFO] Listening to port %v.", port)
	var err error
	n.Listener, err = net.Listen("tcp", ":"+strconv.Itoa(port))
	utils.GetLogger().Printf("[INFO] New listener on node: %v.", n)
	if err != nil {
		return err
	}
	return nil
}

func (n *Cloud) AcceptListener() {
	utils.GetLogger().Println("[INFO] Entering loop to accept clients.")
	for {
		conn, err := n.Listener.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		utils.GetLogger().Printf("[INFO] Accepted connection: %v", conn)

		node := &Node{
			IP: conn.RemoteAddr().String(),
			client: comm.NewClient(conn),
		}
		utils.GetLogger().Printf("[INFO] Connected to a new node: %v", node)
		node.client.AddRequestHandler(createAuthRequestHandler(node, n))
		n.PendingNodes = append(n.PendingNodes, node)
		utils.GetLogger().Printf("[DEBUG] Added node to pending nodes: %v", n.PendingNodes)
		go func(node *Node) {
			node.client.HandleConnection()

			n.Mutex.Lock()
			for _, c := range n.Network.Nodes {
				if c.ID == node.ID {
					utils.GetLogger().Printf("[DEBUG] Node: %v, setting client to nil: %v", c.ID, c.client)
					c.client = nil
				}
			}
			n.Mutex.Unlock()
		}(node)
	}
}

func (n *Node) Ping() (string, error) {
	utils.GetLogger().Println("[INFO] Pinging node.")
	ping, err := n.client.SendMessage("ping", "ping")
	if err != nil {
		return "", err
	}
	return ping[0].(string), err
}

func (r request) PingRequest(ping string) string {
	utils.GetLogger().Printf("[INFO] Handling ping request with string: %v.", ping)
	if ping == "ping" {
		return "pong"
	}
	return ""
}

func (n *Node) Online() bool {
	utils.GetLogger().Println("[DEBUG] Checking if node is online.")
	return n.client != nil
}