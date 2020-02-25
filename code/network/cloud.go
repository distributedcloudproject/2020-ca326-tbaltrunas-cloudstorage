package network

import (
	"cloud/datastore"
	"crypto/rsa"
	"net"
	"sync"
)

type Cloud interface {
	// Listening.
	Listen() error
	ListenOnPort(port int) error
	Accept()
	AcceptUsingListener(listener net.Listener)
	ListenAndAccept() error

	// Config.
	Config() CloudConfig
	SetConfig(config CloudConfig)

	// Network.
	Network() Network
	MyNode() Node
	OnlineNodesNum() int
	NodesNum() int

	// Nodes.
	AddNode(node Node)
	IsNodeOnline(ID string) bool
	GetCloudNode(ID string) *cloudNode

	// Whitelist.
	AddToWhitelist(ID string) error
	Whitelist() []string

	// File
	AddFile(file *datastore.File) error
	Distribute(file datastore.File, numReplicas int, antiAffinity bool) error

	// Events.
	Events() *CloudEvents

	// Saving.
	SavedNetworkState() SavedNetworkState

	// Benchmark state for this cloud.
	BenchmarkState() CloudBenchmarkState
	SetBenchmarkState(benchmarkState CloudBenchmarkState)
}

type CloudEvents struct {
	NodeAdded   func(node Node)
	NodeUpdated func(node Node)
	NodeRemoved func(ID string)

	NodeConnected    func(ID string)
	NodeDisconnected func(ID string)

	WhitelistAdded   func(ID string)
	WhitelistRemoved func(ID string)
}

// Cloud is the client's view of the Network. Contains client-specific information.
type cloud struct {
	network Network
	// NetworkMutex is used only when accessing the Network
	networkMutex sync.RWMutex

	// Nodes maps ID -> cloudNode. It only contains online nodes that we are connected with.
	// This should always include local connection, a cloudNode that corresponds with us.
	Nodes map[string]*cloudNode
	// NodesMutex is used only when accessing the Nodes.
	NodesMutex sync.RWMutex

	// Mutex is used for any other cloud variable.
	Mutex sync.RWMutex

	// Non-authorized connections.
	PendingNodes []*cloudNode

	// Used for events.
	events *CloudEvents

	myNode     Node
	privateKey *rsa.PrivateKey

	listener net.Listener
	Port     int

	config CloudConfig

	benchmarkState CloudBenchmarkState
}

func (c *cloud) Config() CloudConfig {
	return c.config
}

func (c *cloud) SetConfig(config CloudConfig) {
	c.config = config
}

func (c *cloud) Events() *CloudEvents {
	return c.events
}

func (c *cloud) MyNode() Node {
	c.Mutex.RLock()
	defer c.Mutex.RUnlock()
	return c.myNode
}

func (c *cloud) PrivateKey() *rsa.PrivateKey {
	c.Mutex.RLock()
	defer c.Mutex.RUnlock()
	return c.privateKey
}

func (c *cloud) OnlineNodesNum() int {
	c.NodesMutex.RLock()
	defer c.NodesMutex.RUnlock()

	return len(c.Nodes)
}

func (c *cloud) NodesNum() int {
	c.networkMutex.RLock()
	defer c.networkMutex.RUnlock()

	return len(c.network.Nodes)
}

func (c *cloud) Network() Network {
	c.networkMutex.RLock()
	defer c.networkMutex.RUnlock()

	return c.network
}

func (c *cloud) ListenAddress() string {
	if c.listener == nil {
		return ""
	}
	return c.listener.Addr().String()
}
