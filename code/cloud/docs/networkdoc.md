# network
--
    import "cloud/cloud/network"


## Usage

```go
const (
	NetworkInfoMsg    = "NetworkInfo"
	NodeInfoMsg       = "NodeInfo"
	AddNodeMsg        = "AddNode"
	AddToWhitelist    = "AddToWhitelist"
	RemoveToWhitelist = "RemoveFromWhitelist"
)
```
Messages used for basic communication.

```go
const (
	StorageSpaceRemainingMsg = "StorageSpaceRemaining"
	NetworkLatencyMsg        = "NetworkLatency"
)
```
Messages used for benchmark communications.

```go
const (
	CreateDirectoryMsg = "CreateDirectory"
	DeleteDirectoryMsg = "DeleteDirectory"

	AddFileMsg    = "AddFile"
	UpdateFileMsg = "UpdateFile"
	DeleteFileMsg = "DeleteFile"
	MoveFileMsg   = "MoveFile"

	SaveChunkMsg = "SaveChunk"
	GetChunkMsg  = "GetChunk"

	LockFileMsg   = "LockFile"
	UnlockFileMsg = "UnlockFile"
)
```
Messages for data communications.

```go
const (
	AuthMsg = "authenticate_me"
)
```
Messages used for authentication.

```go
const DOWNLOAD_WORKERS = 50
```
The maximum amount of threads to spawn when downloading.

#### func  CleanNetworkPath

```go
func CleanNetworkPath(networkPath string) string
```
CleanNetworkPath cleans the provided path and returns a network-friendly path.
Always starting with a / and only containing forward slashes.

#### func  IsDirEmpty

```go
func IsDirEmpty(name string) (bool, error)
```
IsDirEmpty returns true if the provided directory is empty.

#### func  PublicKeyToID

```go
func PublicKeyToID(key *rsa.PublicKey) (string, error)
```
PublicKeyToID generates a sha256 hash based on the public key.

#### type AuthRequest

```go
type AuthRequest struct {
	ID   string
	IP   string
	Name string
}
```

AuthRequest is sent when attempting to authenticate with a node.

#### type ChunkNodes

```go
type ChunkNodes map[datastore.ChunkID][]string
```


#### type Cloud

```go
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
```


#### func  BootstrapToNetwork

```go
func BootstrapToNetwork(bootstrapIP string, myNode Node, privateKey *rsa.PrivateKey, config CloudConfig) (Cloud, error)
```

#### func  LoadNetwork

```go
func LoadNetwork(s SavedNetworkState) Cloud
```
LoadNetwork uses a SavedNetworkState to connect to the network, if it's up. If
the network is offline, it will bring it back online.

#### func  SetupNetwork

```go
func SetupNetwork(network Network, myNode Node, privateKey *rsa.PrivateKey) Cloud
```

#### func  SetupNetworkWithConfig

```go
func SetupNetworkWithConfig(network Network, myNode Node, privateKey *rsa.PrivateKey, config CloudConfig) Cloud
```

#### type CloudBenchmarkState

```go
type CloudBenchmarkState struct {
	StorageSpaceUsed uint64 // in bytes, how much storage is used already.
}
```

benchmarks for this cloud's node.

#### type CloudConfig

```go
type CloudConfig struct {
	// FileStorageDir is a file path to a directory where user files should be stored on this node.
	FileStorageDir string

	// FileStorageCapacity is the maximum amount of user data that should be stored on this node, in bytes.
	// If 0, the node's available disk capacity (under the FileStorageDir path) will be taken as the storage capacity.
	// If -1, no storage will be allowed on the node.
	FileStorageCapacity int64

	// FileChunkSize controls into how many bytes a file should be chunked in.
	FileChunkSize int
}
```


#### type CloudEvents

```go
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
```


#### type DownloadEvent

```go
type DownloadEvent int
```


```go
const (
	DownloadCompleted DownloadEvent = iota
	InfoRetrieved
)
```

#### type DownloadManager

```go
type DownloadManager struct {
	Cloud *cloud
}
```


#### func (*DownloadManager) Completed

```go
func (m *DownloadManager) Completed() int
```

#### func (*DownloadManager) DownloadFile

```go
func (m *DownloadManager) DownloadFile(cloudPath string, localPath string) error
```

#### func (*DownloadManager) Downloading

```go
func (m *DownloadManager) Downloading() int
```

#### func (*DownloadManager) QueueDownload

```go
func (m *DownloadManager) QueueDownload(cloudPath, localPath string, OnComplete func(event DownloadEvent)) *DownloadQueue
```

#### func (*DownloadManager) Queued

```go
func (m *DownloadManager) Queued() int
```

#### type DownloadQueue

```go
type DownloadQueue struct {
	CloudPath       string
	LocalPath       string
	ChunkDownloaded []bool
	Completed       bool
	OnEvent         func(event DownloadEvent)
}
```


#### type FileNodes

```go
type FileNodes map[datastore.FileID][]string
```


#### type Network

```go
type Network struct {
	Name  string
	Nodes []Node

	// Require authentication for the network. Authentication verifies that Node ID belongs to the public key.
	RequireAuth bool

	// Enable whitelist for the network. If enabled, Node ID has to be whitelisted before joining the network.
	Whitelist bool

	// List of node IDs that are permitted to enter the network.
	WhitelistIDs []string

	RootFolder *NetworkFolder

	// ChunkNodes maps chunk ID's to the Nodes (Node ID's) that contain that chunk.
	// This way we can keep track of which nodes contain which chunks.
	// And make decisions about the chunk requests to perform.
	// In the future this scheme might change, for example, with each node knowing only about its own chunks.
	ChunkNodes ChunkNodes

	// FileNodes maps file ID's to the Nodes that contain the whole file. Those nodes are syncing the whole file all the
	// time.
	FileNodes FileNodes
}
```

Network is the general info of the network. Each node would have the same
presentation of Network.

#### func (*Network) GetFile

```go
func (n *Network) GetFile(file string) (*datastore.File, error)
```
GetFile retrieves the metadata of a file given it's path.

#### func (*Network) GetFiles

```go
func (n *Network) GetFiles() []*datastore.File
```
GetFiles retrieve the files in the network.

#### func (*Network) GetFolder

```go
func (n *Network) GetFolder(folder string) (*NetworkFolder, error)
```
GetFolder retrieves the folder for the given path.

#### func (*Network) GetFolders

```go
func (n *Network) GetFolders() []*NetworkFolder
```
GetFolders retrieve the folders in the network.

#### func (*Network) NodeByID

```go
func (n *Network) NodeByID(ID string) (node Node, found bool)
```

#### type NetworkFolder

```go
type NetworkFolder struct {
	Name       string
	SubFolders []*NetworkFolder

	// Files is a list of files in current folder on the Cloud.
	Files datastore.DataStore
}
```


#### type Node

```go
type Node struct {
	// Unique ID of the Node.
	ID string

	// Represented in ip:port format.
	// Example: 127.0.0.1:8081
	IP string

	// Display name of the node.
	Name string

	// Public key of the node.
	PublicKey crypto.PublicKey
}
```

Node is a global representation of any node. Each network will have the same
view of the node.

#### type NodeBenchmark

```go
type NodeBenchmark struct {
	ID                    string
	StorageSpaceRemaining uint64
	Latency               time.Duration
}
```

NodeBenchmark represents a set of benchmarks for a node given by ID.

#### type Response

```go
type Response struct {
	Returns []interface{}
	Error   error
	Node    *cloudNode
}
```


#### type SaveChunkRequest

```go
type SaveChunkRequest struct {
	FilePath string
	Chunk    datastore.Chunk // chunk metadata

	Contents []byte // chunk bytes
}
```


#### type SavedNetworkState

```go
type SavedNetworkState struct {
	Network Network
	Config  CloudConfig

	MyNode     Node
	PrivateKey *rsa.PrivateKey

	FileStorage map[string]datastore.FileStore
	FileSyncs   []*datastore.SyncFileStore
	FolderSyncs []fileSync
}
```

SavedNetworkState contains information to re-spawn a network from an offline
state.
