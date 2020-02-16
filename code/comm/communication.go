package comm

import (
	"cloud/utils"
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"net"
	"reflect"
	"sync"
	"sync/atomic"
)

// message is used to keep track of sent requests/messages and retrieving the response.
type message struct {
	// The data that was received as per response.
	value []byte
	err error

	// The wait group allows us to block the sending thread until a response is received.
	wg sync.WaitGroup
}

type RequestHandler func(message string) interface{}

// Client represents a connection to another node.
type Client interface {
	RegisterRequest(message string, f interface{})
	AddRequestHandler(handler RequestHandler)

	Address() string

	SendMessage(msg string, data ...interface{}) ([]interface{}, error)
	HandleConnection() error
	Close() error
}

type client struct {
	conn net.Conn

	// messages is used to retrieve responses from a request.
	// Lock the mutex when accessing the map.
	messages map[uint32]*message // TODO: timeout
	messagesMutex sync.Mutex

	// Access msgID atomically.
	msgID uint32

	// Used to lock when writing to the connection. While conn.Write is thread safe, it does not guarantee that
	// all of the data will be sent at once. Since we rely on packets to come in order, it is important to lock until
	// all of the data is written.
	writeMutex sync.Mutex

	requests map[string]interface{}
	requestsHandlers []RequestHandler
	requestsMutex sync.RWMutex
}

// NewClient creates a new client with an existing network connection.
func NewClient(conn net.Conn) Client {
	utils.GetLogger().Printf("Creating new client from connection: %v.", conn)
	client := &client{}

	client.conn = conn
	client.messages = make(map[uint32]*message)
	client.requests = make(map[string]interface{})

	utils.GetLogger().Printf("Created new client: %v.", client)
	return client
}

// NewClientDial creates a new client by dialing the ip and creating a new socket connection.
func NewClientDial(address string) (Client, error) {
	utils.GetLogger().Printf("Creating a new client from a dial to address: %v.", address)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	return NewClient(conn), nil
}

func (c *client) Address() string {
	return c.conn.RemoteAddr().String()
}

func (c *client) Close() error {
	return c.conn.Close()
}

func (c *client) RegisterRequest(message string, f interface{}) {
	c.requestsMutex.Lock()
	defer c.requestsMutex.Unlock()
	utils.GetLogger().Printf("Registering request for message: %v, with interface: %v.", message, f)
	c.requests[message] = f
}

func (c *client) AddRequestHandler(handler RequestHandler) {
	c.requestsMutex.Lock()
	defer c.requestsMutex.Unlock()
	utils.GetLogger().Printf("Adding request handler: %v.", handler)
	c.requestsHandlers = append(c.requestsHandlers, handler)
}

// SendMessage sends a request with the msg and the data passed. Returns a list of arguments that were returned.
func (c *client) SendMessage(msg string, data ...interface{}) ([]interface{}, error) {
	utils.GetLogger().Printf("Sending message: %v, with data: %v.", msg, data)
	id := atomic.AddUint32(&c.msgID, 1)
	utils.GetLogger().Printf("Computed message ID: %v.", id)

	// The headers take up 9 bytes.
	// | Is a response (1) | Message ID (4) | Message Length (4) |
	// The first byte will be used as a boolean to see if it's a response.
	// The next 4 bytes will hold the message ID (used to link the response back).
	// The next 4 bytes will hold the message length.
	buffer := make([]byte, 9)
	binary.LittleEndian.PutUint32(buffer[1:5], id)

	// Add the message/function name to the buffer, finished by the \000.
	buffer = append(buffer, []byte(msg)...)
	buffer = append(buffer, '\000')

	// Encode our data that we want to send using gob.
	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)
	for i := range data {
		err := e.Encode(&data[i])
		if err != nil {
			return nil, err
		}
	}
	// Copy over the encoded bytes buffer into our buffer variable.
	buffer = append(buffer, b.Bytes()...)

	// Put the buffer length into [5:9] bytes. -9 because the message length should not include the headers(message ID
	// and message length), which the buffer contains.
	binary.LittleEndian.PutUint32(buffer[5:9], uint32(len(buffer)-9))
	utils.GetLogger().Printf("Prepared message buffer (isResponse, ID, dataLength): %v.", buffer[:9])

	// Place our message into the map, so that it can be used when receiving responses.
	m := &message{
		wg: sync.WaitGroup{},
	}
	utils.GetLogger().Printf("Created message: %v.", m)
	c.messagesMutex.Lock()
	c.messages[id] = m
	utils.GetLogger().Printf("Added message to map: %v.", c.messages)
	c.messagesMutex.Unlock()

	m.wg.Add(1)

	// Write the buffer to the socket connection.
	utils.GetLogger().Println("Writing buffer to socket.")
	written := 0
	c.writeMutex.Lock()
	for written < len(buffer) {
		n, err := c.conn.Write(buffer[written:])
		if err != nil {
			return nil, err
		}
		written += n
	}
	c.writeMutex.Unlock()
	utils.GetLogger().Println("Finished writing buffer to socket.")

	// Waitgroup will be released when there's a response to the request.
	utils.GetLogger().Println("Blocking until receive response to request.")
	m.wg.Wait()
	utils.GetLogger().Println("Received response to request.")

	// Decode the response from gob into interface{} values. The interface{} then can be casted onto their original
	// type.
	d := gob.NewDecoder(bytes.NewBuffer(m.value))
	vars := make([]interface{}, 0)
	var err error
	for err == nil {
		var a interface{}
		err = d.Decode(&a)
		if err == nil {
			vars = append(vars, a)
		}
	}
	utils.GetLogger().Println("Extracted variables from message.")

	if err != io.EOF {
		return nil, err
	}

	return vars, nil
}

func (c *client) HandleConnection() error {
	utils.GetLogger().Println("Starting handling connection loop.")
	for {
		headerBuffer := make([]byte, 9)
		utils.GetLogger().Println("Reading header from socket.")
		_, err := c.conn.Read(headerBuffer)
		if err != nil {
			if err == io.EOF || !err.(net.Error).Temporary() {
				return err
			}
			continue
		}
		utils.GetLogger().Printf("Read header into buffer: %v.", headerBuffer)

		response := false
		if headerBuffer[0] == 1 {
			response = true
		}
		messageID := binary.LittleEndian.Uint32(headerBuffer[1:5])
		messageLength := int(binary.LittleEndian.Uint32(headerBuffer[5:9]))
		utils.GetLogger().Printf("Extracted from header isResponse: %v, messageID: %v, messageLength: %v.", 
								 response, messageID, messageLength)

		buffer := make([]byte, messageLength)
		totalRead := 0

		utils.GetLogger().Println("Reading contents from socket.")
		for totalRead < messageLength {
			read, err := c.conn.Read(buffer[totalRead:])
			if err != nil {
				return err
			}
			totalRead += read
		}
		utils.GetLogger().Println("Finished reading contents into buffer.")

		// Once the data is retrieved, process it in another thread so that we can continue receiving data.
		utils.GetLogger().Println("Passing data processing to another thread.")
		go func() {
			err := c.processRequest(response, messageID, buffer)
			if err != nil {
				fmt.Println(err)
			}
		}()
	}
}

func (c *client) processRequest(response bool, messageID uint32, data []byte) error {
	utils.GetLogger().Printf("Processing request isResponse: %v, messageID: %v.", response, messageID)
	if response {
		utils.GetLogger().Println("Processing a request of response type.")
		c.messagesMutex.Lock()
		utils.GetLogger().Printf("Removing message from messages map: %v.", c.messages)
		message, ok := c.messages[messageID]
		if ok {
			delete(c.messages, messageID)
			message.value = data
			message.wg.Done()
		}
		utils.GetLogger().Printf("Updated message map: %v.", c.messages)
		c.messagesMutex.Unlock()

		return nil
	}

	utils.GetLogger().Println("Processing request of non-response type.")
	// Extract the function name from the request.
	index := bytes.IndexByte(data, '\000')
	if index == -1 {
		fmt.Println(index, data)
		return errors.New("request incorrectly formed - could not extract function name")
	}
	funcName := string(data[:index])
	utils.GetLogger().Printf("Extracted function name: %v.", funcName)
	c.requestsMutex.RLock()
	request, ok := c.requests[funcName]
	utils.GetLogger().Printf("Got request handler for function: %v.", request)
	if !ok {
		for i := 0; i < len(c.requestsHandlers); i++ {
			request = c.requestsHandlers[i](funcName)
			if request != nil {
				ok = true
				break
			}
		}
	}
	c.requestsMutex.RUnlock()

	if !ok {
		return errors.New("function " + funcName + " is not registered.")
	}

	// Decode the buffer into variables.
	utils.GetLogger().Println("Extracting variables from data.")
	d := gob.NewDecoder(bytes.NewBuffer(data[index+1:]))
	vars := make([]reflect.Value, 0)
	var err error
	for err == nil {
		var a interface{}
		err = d.Decode(&a)
		if err == nil {
			vars = append(vars, reflect.ValueOf(a))
		}
	}
	if err != io.EOF {
		return err
	}
	utils.GetLogger().Println("Finished extracting variables.")
	returnVars := reflect.ValueOf(request).Call(vars)
	utils.GetLogger().Printf("Return values of request handler with vars: %v.", returnVars)

	// Encode the return variables and send them as reply.
	utils.GetLogger().Println("Encoding returned values.")
	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)
	for i := range returnVars {
		d := returnVars[i].Interface()
		err := e.Encode(&d)
		if err != nil {
			return err
		}
	}
	utils.GetLogger().Println("Finished encoding return values.")

	buffer := make([]byte, 9+b.Len())
	buffer[0] = 1 // This is a response.
	binary.LittleEndian.PutUint32(buffer[1:5], messageID)
	binary.LittleEndian.PutUint32(buffer[5:9], uint32(b.Len()))
	copy(buffer[9:], b.Bytes())
	utils.GetLogger().Printf("Created a response buffer (isResponse, messageID, length): %v.", buffer[:9])

	written := 0
	c.writeMutex.Lock()
	utils.GetLogger().Println("Writing response to socket.")
	for written < len(buffer) {
		n, err := c.conn.Write(buffer[written:])
		if err != nil {
			return err
		}
		written += n
	}
	utils.GetLogger().Println("Finished writing response to socket.")
	c.writeMutex.Unlock()
	return nil
}
