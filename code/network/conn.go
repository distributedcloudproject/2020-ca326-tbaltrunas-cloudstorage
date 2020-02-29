package network

import (
	"cloud/comm"
	"cloud/datastore"
	"cloud/utils"
	"crypto/rsa"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

type CloudConfig struct {
	// FileStorageDir is a file path to a directory where user files should be stored on this node.
	FileStorageDir string

	// FileStorageCapacity is the maximum amount of user data that should be stored on this node, in bytes.
	// If 0, the node's available disk capacity (under the FileStorageDir path) will be taken as the storage capacity.
	// If -1, no storage will be allowed on the node.
	FileStorageCapacity int64
}

// ConnectToNode establishes a connection to a node with that ID. Will return error if a connection could not be
// established or if the node does not exist in the network. Will not return an error if a connection is already
// established. Updates c.Nodes[ID] to point to the cloudNode.
func (c *cloud) ConnectToNode(ID string) error {
	utils.GetLogger().Printf("[INFO] Cloud: %v, connecting to node: %v", c, ID)

	if c.hasCloudNode(ID) {
		// Connection is already established.
		return nil
	}

	// Check that there is a node with corresponding ID.
	if n, found := c.NodeByID(ID); found && n.IP != "" {
		utils.GetLogger().Printf("[DEBUG] Connecting to a non-me node with nil client: %v.", n)

		// Initialize the connection and add the auth handlers.
		client, err := comm.NewClientDial(n.IP, c.PrivateKey())
		if err != nil {
			return err
		}

		// Create a cloudNode that corresponds with the connection.
		node := &cloudNode{
			ID:     ID,
			client: client,
		}

		client.AddRequestHandler(createAuthRequestHandler(node, c))

		// Handle the connection.
		go c.handleCloudNodeConnection(node)

		// Register any request handlers.
		c.addRequestHandlers(node)

		// Authenticate with the target node.
		success, err := node.Authenticate(c.MyNode())
		if !success || err != nil {
			utils.GetLogger().Printf("[INFO] Failed to authenticate with node %v [success: %v | err: %v]", n, success, err)
			node.client.Close()
			if err != nil {
				return err
			}
			return errors.New("node refused to authenticate")
		}

		// If another connection has been established with the target node between checking and connecting, drop the
		// current connection.
		if !c.addCloudNode(ID, node) {
			node.client.Close()
		}
	}
	return nil
}

func BootstrapToNetwork(bootstrapIP string, myNode Node, privateKey *rsa.PrivateKey) (Cloud, error) {
	utils.GetLogger().Printf("[INFO] Bootstrapping with ip: %v, and node: %v.", bootstrapIP, myNode)

	myNode.PublicKey = privateKey.PublicKey
	myNode.ID, _ = PublicKeyToID(&privateKey.PublicKey)

	// Create the cloud object.
	cloud := &cloud{
		Nodes:      make(map[string]*cloudNode),
		events:     &CloudEvents{},
		myNode:     myNode,
		privateKey: privateKey,
		Port:       0,
	}
	ips := strings.Split(myNode.IP, ":")
	if len(ips) > 0 {
		cloud.Port, _ = strconv.Atoi(ips[len(ips)-1])
	}

	cloud.Nodes[myNode.ID] = &cloudNode{
		ID:     myNode.ID,
		client: comm.NewLocalClient(),
	}
	cloud.addRequestHandlers(cloud.Nodes[myNode.ID])

	// Establish connection with the target.
	client, err := comm.NewClientDial(bootstrapIP, privateKey)
	if err != nil {
		return nil, err
	}

	// Create a cloudNode representation of the bootstrap using temporary ID.
	bsNode := &cloudNode{
		ID:     "temp",
		client: client,
	}
	bsNode.client.AddRequestHandler(createAuthRequestHandler(bsNode, cloud))
	go cloud.handleCloudNodeConnection(bsNode)
	cloud.addRequestHandlers(bsNode)

	// Authenticate with the bootstrap node.
	success, err := bsNode.Authenticate(myNode)
	if err != nil {
		bsNode.client.Close()
		return nil, err
	}
	if !success {
		bsNode.client.Close()
		return nil, errors.New("bootstrap node refused to authenticate")
	}

	nodeInfo, err := bsNode.NodeInfo()
	if err != nil {
		bsNode.client.Close()
		return nil, err
	}

	bsNode.ID = nodeInfo.ID
	utils.GetLogger().Printf("[INFO] Bootstrap node ID: %v", nodeInfo.ID)

	cloud.addCloudNode(nodeInfo.ID, bsNode)

	// Retrieve Network Info.
	network, err := bsNode.NetworkInfo()
	if err != nil {
		bsNode.client.Close()
		return nil, err
	}

	cloud.network = network
	utils.GetLogger().Printf("[INFO] Retrieved network info: %v", network)

	// Connect to all of the other nodes.
	go func() {
		for i := range network.Nodes {
			go cloud.ConnectToNode(network.Nodes[i].ID)
		}
	}()

	return cloud, nil
}

func SetupNetwork(network Network, myNode Node, privateKey *rsa.PrivateKey) Cloud {
	utils.GetLogger().Printf("[INFO] Setting up network with name: %v, and initial name: %v.", network.Name, myNode.Name)

	myNode.PublicKey = privateKey.PublicKey
	myNode.ID, _ = PublicKeyToID(&privateKey.PublicKey)

	if network.ChunkNodes == nil {
		network.ChunkNodes = make(map[datastore.ChunkID][]string)
	}

	cloud := &cloud{
		network:    network,
		events:     &CloudEvents{},
		Nodes:      make(map[string]*cloudNode),
		myNode:     myNode,
		privateKey: privateKey,
		Port:       0,
	}
	ips := strings.Split(myNode.IP, ":")
	if len(ips) > 0 {
		cloud.Port, _ = strconv.Atoi(ips[len(ips)-1])
	}
	// If our node is not present in the network; add it.
	if _, ok := cloud.NodeByID(myNode.ID); !ok {
		cloud.network.Nodes = append(cloud.network.Nodes, myNode)
	}
	cloud.Nodes[myNode.ID] = &cloudNode{
		ID:     myNode.ID,
		client: comm.NewLocalClient(),
	}
	cloud.addRequestHandlers(cloud.Nodes[myNode.ID])
	return cloud
}
func SetupNetworkWithConfig(network Network, myNode Node, privateKey *rsa.PrivateKey, config CloudConfig) Cloud {
	c := SetupNetwork(network, myNode, privateKey)
	c.SetConfig(config)
	return c
}

func (c *cloud) ListenAndAccept() error {
	err := c.Listen()
	if err != nil {
		return err
	}
	c.Accept()
	return nil
}

func (c *cloud) ListenOnPort(port int) error {
	c.Port = port
	return c.Listen()
}

func (c *cloud) Listen() error {
	utils.GetLogger().Printf("[INFO] Listening to port %v.", c.Port)
	var err error
	c.listener, err = net.Listen("tcp", ":"+strconv.Itoa(c.Port))

	// Auto-assigned port, update our node IP to reflect to port.
	if c.Port == 0 {
		s := strings.Split(c.listener.Addr().String(), ":")
		newPort, _ := strconv.Atoi(s[len(s)-1])

		myIP := c.myNode.IP
		if len(myIP) == 0 || myIP[0] == ':' {
			myIP = ":" + strconv.Itoa(newPort)
		} else {
			ips := strings.Split(myIP, ":")
			myIP = ips[0] + ":" + strconv.Itoa(newPort)
		}
		c.myNode.IP = myIP
		c.AddNode(c.myNode)
	}
	utils.GetLogger().Printf("[INFO] New listener on node: %v.", c.MyNode().ID)
	if err != nil {
		return err
	}
	return nil
}

func (c *cloud) Accept() {
	utils.GetLogger().Println("[INFO] Entering loop to accept clients.")
	for {
		conn, err := c.listener.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		utils.GetLogger().Printf("[INFO] Accepted connection: %v", conn)

		client, err := comm.NewServerClient(conn, c.PrivateKey())
		if err != nil {
			utils.GetLogger().Printf("[INFO] Could not create server client with %v: %v", conn.RemoteAddr(), err)
			continue
		}
		node := &cloudNode{
			client: client,
		}
		utils.GetLogger().Printf("[INFO] Connected to a new node: %v", node)

		node.client.AddRequestHandler(createAuthRequestHandler(node, c))
		c.Mutex.Lock()
		c.PendingNodes = append(c.PendingNodes, node)
		c.Mutex.Unlock()

		utils.GetLogger().Printf("[DEBUG] Added node to pending nodes: %v", c.PendingNodes)

		go c.handleCloudNodeConnection(node)
	}
}

func (c *cloud) AcceptUsingListener(listener net.Listener) {
	c.listener = listener
	c.Accept()
}

func (c *cloud) handleCloudNodeConnection(n *cloudNode) {
	n.client.HandleConnection()

	// Remove connection from our node list.
	if cn := c.GetCloudNode(n.ID); cn != nil && cn.client == n.client {
		c.removeCloudNode(n.ID)
		return
	}

	// Remove from pendingNodes if it's there.
	c.removePendingNode(n.client)
}
