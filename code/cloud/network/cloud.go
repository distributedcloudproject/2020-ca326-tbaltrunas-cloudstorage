package network

import (
	"cloud/datastore"
	"crypto/rsa"
	"github.com/fsnotify/fsnotify"
	"net"
	"os"
	"sync"
	"time"
)

type Cloud interface {
	// Listen creates a listener on the set cloud's port. Does not handle incoming connections.
	Listen() error
	// ListenOnPort creates a listener on the specified port. Does not handle incoming connections.
	ListenOnPort(port int) error
	// Accept handles incoming connections. A listener has to be created first using Listen or ListenOnPort.
	Accept()
	// AcceptUsingListener handles incoming connections using a provided listener.
	AcceptUsingListener(listener net.Listener)
	// ListenAndAccept creates a listener and handles incoming connections.
	ListenAndAccept() error

	// Config retrieves the config that was set for the cloud instance.
	Config() CloudConfig
	// SetConfig sets the cloud's instance config.
	SetConfig(config CloudConfig)

	// Network retrieves the Network information. This information is synced between all nodes.
	Network() Network
	// MyNode retrieves the Node data for the local node.
	MyNode() Node
	// OnlineNodesNum returns the amount of nodes that are online.
	OnlineNodesNum() int
	// NodesNum returns the amount of nodes that are present in the network/
	NodesNum() int
	// DownloadManager returns an instance of the download manager, which can be used to queue downloads.
	DownloadManager() *DownloadManager

	// AddNode adds a node to the network. The request is sent to all of the nodes in the network.
	AddNode(node Node)
	// IsNodeOnline returns if a specified node by an ID is online on the network/
	IsNodeOnline(ID string) bool
	// GetCloudNode returns an online instance of the specified node ID. If the node is not online, or is not present
	// in the network, it will return nil.
	GetCloudNode(ID string) *cloudNode
	// NodeByID returns the node instance for ID and if that node was found in the network.
	NodeByID(ID string) (node Node, found bool)

	// AddToWhitelist adds a node ID to the network's whitelist. Node ID has to be added to the whitelist before joining
	// the network.
	AddToWhitelist(ID string) error
	// Whitelist returns a list of node IDs that are whitelisted in the network.
	Whitelist() []string

	// GetFolder retrieves virtual folder on the cloud. It will create the folder if one does not exist.
	GetFolder(path string) (*NetworkFolder, error)
	// GetFile retrieves the metadata of a file on the cloud.
	GetFile(file string) (*datastore.File, error)
	// GetFiles retrieves all metadata files on the cloud.
	GetFiles() []*datastore.File
	// GetFolders retrieves all folders on the cloud.
	GetFolders() []*NetworkFolder
	// DistributeChunk calculates where it should distribute the chunk on the cloud and sends the data over.
	DistributeChunk(cloudPath string, store datastore.FileStore, chunkID datastore.ChunkID) error
	// CreateDirectory creates a directory on the cloud.
	CreateDirectory(folderPath string) error
	// DeleteDirectory deletes a directory from the cloud. The directory must be empty.
	DeleteDirectory(folderPath string) error
	// AddFile creates a new file on the cloud, copying the file from localpath.
	AddFile(file *datastore.File, filepath string, localpath string) error
	// AddFileInPlace creates a new file on the cloud, using the localpath as it's store file. Instead of making another
	// copy of the file in the storage directory, it will use the localpath to store the file on the local node.
	AddFileInPlace(file *datastore.File, filepath string, localpath string) error
	// AddFileMetadata creates a new file on the cloud without sending over any data. Chunks will have to be distributed
	// using DistributeChunk.
	AddFileMetadata(file *datastore.File, filepath string) error
	// UpdateFile updates the file's metadata. The node calling UpdateFile needs to have the data for any new chunks.
	// File lock is required.
	UpdateFile(file *datastore.File, filepath string) error
	// DeleteFile deletes a file from the cloud. File lock is required.
	DeleteFile(filepath string) error
	// MoveFile moves a file on the cloud to a new path. File lock is required for old and new file paths.
	MoveFile(filepath string, newFilepath string) error
	// LockFile creates a file lock on the specified path. Returns true if one could be acquired.
	// File lock is required for manipulating files. Prevents from multiple nodes changing the same file at the same
	// time, which could lose data.
	LockFile(path string) bool
	// UnlockFile releases the file lock for the path. Only the node that has the file lock for that file can unlock it.
	UnlockFile(path string)
	// SyncFile creates a sync between a cloud file and local file. Those files will be kept the same. Any changes to
	// the cloud file will be reflected in the local file, and any changes in the local file will be reflected in the
	// cloud file. Uses fsnotify to monitor for changes in the local file.
	// If the cloud file does not exist, one will be created from the local file.
	// If the local file does not exist, one will be created from the cloud file.
	// If both exist, they have to be the same.
	SyncFile(cloudPath string, localPath string) error
	// SyncFolder creates a sync between a cloud folder and a local folder. This is similar to SyncFile, but instead of
	// syncing individual files, it will sync the whole folder. The local folder has to be empty, the cloud folder has to
	// exist.
	SyncFolder(cloudPath string, localPath string) error
	// Distribute calculates what nodes to split the data to and replicates the data to those nodes.
	Distribute(cloudPath string, file datastore.File, numReplicas int, antiAffinity bool) error

	// Events returns a cloud event instance, which can be used to set event hooks.
	Events() *CloudEvents

	// SavedNetworkState returns a saved instance of the cloud.
	SavedNetworkState() SavedNetworkState

	// BenchmarkState returns benchmark information for this cloud's node.
	BenchmarkState() CloudBenchmarkState
	// SetBenchmarkState sets the benchmark information for this cloud's node.
	SetBenchmarkState(benchmarkState CloudBenchmarkState)
}

type CloudEvents struct {
	// NodeAdded is called when a new node is added to the network.
	NodeAdded func(node Node)
	// NodeUpdated is called when an existing node is updated.
	NodeUpdated func(node Node)
	// NodeRemoved is called when a node is kicked out of the network.
	NodeRemoved func(ID string)

	// NodeConnected is called when a node connects to the network.
	NodeConnected func(ID string)
	// NodeDisconnected is called when a node disconnects from the network.
	NodeDisconnected func(ID string)

	// WhitelistAdded is called when a new whitelist ID is added on the network.
	WhitelistAdded func(ID string)
	// WhitelistRemoved is called when a whitelist ID is removed from the network.
	WhitelistRemoved func(ID string)
}

type fileSync struct {
	CloudPath    string
	LocalPath    string
	LastEditTime time.Time
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

	// Locks a file (full path) to a node ID. If a path is locked, only the given node ID may interact
	// with the file. This is to prevent race conditions.
	fileLocks     map[string]string
	fileLockMutex sync.RWMutex

	// Mutex is used for any other cloud variable.
	Mutex sync.RWMutex

	// Local storage. Maps file path to the store.
	fileStorage      map[string]datastore.FileStore
	fileStorageMutex sync.RWMutex

	downloadManager *DownloadManager

	fileSyncs   []fileSync
	folderSyncs []fileSync
	watcher     *fsnotify.Watcher

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

func (c *cloud) DownloadManager() *DownloadManager {
	return c.downloadManager
}

func (c *cloud) FileStore(cloudPath string) datastore.FileStore {
	c.fileStorageMutex.RLock()
	defer c.fileStorageMutex.RUnlock()
	return c.fileStorage[cloudPath]
}

func (c *cloud) Config() CloudConfig {
	return c.config
}

func (c *cloud) SetConfig(config CloudConfig) {
	c.config = config
	os.MkdirAll(c.config.FileStorageDir, os.ModeDir)
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
