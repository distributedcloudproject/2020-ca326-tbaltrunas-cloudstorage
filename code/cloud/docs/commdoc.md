# comm
--
    import "cloud/cloud/comm"


## Usage

#### type Client

```go
type Client interface {
	RegisterRequest(message string, f interface{})
	AddRequestHandler(handler RequestHandler)

	Address() string

	SendMessage(msg string, data ...interface{}) ([]interface{}, error)
	HandleConnection() error
	Close() error
	PublicKey() *rsa.PublicKey
}
```

Client represents a connection to another node.

#### func  NewClient

```go
func NewClient(conn net.Conn, key *rsa.PrivateKey) (Client, error)
```
NewClient creates a new client with an existing network connection.

#### func  NewClientDial

```go
func NewClientDial(address string, key *rsa.PrivateKey) (Client, error)
```
NewClientDial creates a new client by dialing the ip and creating a new socket
connection.

#### func  NewLocalClient

```go
func NewLocalClient() Client
```

#### func  NewServerClient

```go
func NewServerClient(conn net.Conn, key *rsa.PrivateKey) (Client, error)
```
NewServerClient creates a new server client with an existing network connection.
Server Client is the client that accepted the connection, instead of the one
that initiated it. Server Client will generate the master key.

#### type RequestHandler

```go
type RequestHandler func(message string) interface{}
```
