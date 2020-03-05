package comm

import (
	"bytes"
	"cloud/utils"
	cipher2 "crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"net"
	"time"
	"reflect"
	"sync"
	"sync/atomic"
)

var (
	msgTimeout = 30 * time.Second
)

// message is used to keep track of sent requests/messages and retrieving the response.
type message struct {
	// The data that was received as per response.
	value []byte
	err   error

	// The channel allows us to block the sending thread until a response is received.
	ch chan struct{}
}

type RequestHandler func(message string) interface{}

var errorInterface = reflect.TypeOf((*error)(nil)).Elem()

type commError struct {
	Error string
}

func init() {
	gob.Register(commError{})
}

// Client represents a connection to another node.
type Client interface {
	RegisterRequest(message string, f interface{})
	AddRequestHandler(handler RequestHandler)

	Address() string

	SendMessage(msg string, data ...interface{}) ([]interface{}, error)
	HandleConnection() error
	Close() error
	PublicKey() *rsa.PublicKey
}

type client struct {
	conn net.Conn

	// messages is used to retrieve responses from a request.
	// Lock the mutex when accessing the map.
	messages      map[uint32]*message // TODO: timeout
	messagesMutex sync.Mutex

	// Access msgID atomically.
	msgID uint32

	// Used to lock when writing to the connection. While conn.Write is thread safe, it does not guarantee that
	// all of the data will be sent at once. Since we rely on packets to come in order, it is important to lock until
	// all of the data is written.
	writeMutex sync.Mutex

	requests         map[string]interface{}
	requestsHandlers []RequestHandler
	requestsMutex    sync.RWMutex

	// Our private key, used for decryption.
	privateKey *rsa.PrivateKey

	// Their public key, used for encryption.
	publicKey *rsa.PublicKey

	// Master key used for symmetric encryption/decryption.
	masterKey []byte
	cipher    cipher2.AEAD
}

// NewClientDial creates a new client by dialing the ip and creating a new socket connection.
func NewClientDial(address string, key *rsa.PrivateKey) (Client, error) {
	utils.GetLogger().Printf("[DEBUG] Creating a new client from a dial to address: %v.", address)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	return NewClient(conn, key)
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
	utils.GetLogger().Printf("[DEBUG] Registering request for message: %v, with interface: %v.", message, f)
	c.requests[message] = f
}

func (c *client) AddRequestHandler(handler RequestHandler) {
	c.requestsMutex.Lock()
	defer c.requestsMutex.Unlock()
	utils.GetLogger().Printf("[DEBUG] Adding request handler: %v.", handler)
	c.requestsHandlers = append(c.requestsHandlers, handler)
}

func (c *client) PublicKey() *rsa.PublicKey {
	return c.publicKey
}

// SendMessage sends a request with the msg and the data passed. Returns a list of arguments that were returned.
func (c *client) SendMessage(msg string, data ...interface{}) ([]interface{}, error) {
	utils.GetLogger().Printf("[DEBUG] Sending message: %v.", msg)
	id := atomic.AddUint32(&c.msgID, 1)
	utils.GetLogger().Printf("[DEBUG] Computed message ID: %v.", id)

	// The headers take up 9 bytes.
	// | Is a response (1) | Message ID (4) | Message Length (4) |
	// The first byte will be used as a boolean to see if it's a response.
	// The next 4 bytes will hold the message ID (used to link the response back).
	// The next 4 bytes will hold the message length.
	buffer := make([]byte, 9)
	binary.LittleEndian.PutUint32(buffer[1:5], id)

	// Add the message/function name to the buffer, finished by the \000.
	b := bytes.Buffer{}
	b.Write([]byte(msg))
	b.WriteRune('\000')

	// Encode our data that we want to send using gob.
	e := gob.NewEncoder(&b)
	for i := range data {
		err := e.Encode(&data[i])
		if err != nil {
			return nil, err
		}
	}

	// Encrypt the data using symmetric key.
	nonce := make([]byte, c.cipher.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	encryptedBuffer := c.cipher.Seal(nonce, nonce, b.Bytes(), nil)
	buffer = append(buffer, encryptedBuffer...)

	// Put the buffer length into [5:9] bytes. -9 because the message length should not include the headers(message ID
	// and message length), which the buffer contains.
	binary.LittleEndian.PutUint32(buffer[5:9], uint32(len(buffer)-9))
	utils.GetLogger().Printf("[DEBUG] Prepared message buffer (isResponse, ID, dataLength): %v.", buffer[:9])

	// Place our message into the map, so that it can be used when receiving responses.
	m := &message{
		ch: make(chan struct{}),
	}
	defer close(m.ch)

	utils.GetLogger().Printf("[DEBUG] Created message: %v.", m)
	c.messagesMutex.Lock()
	c.messages[id] = m
	utils.GetLogger().Printf("[DEBUG] Added message to map: %v.", c.messages)
	c.messagesMutex.Unlock()

	// Write the buffer to the socket connection.
	utils.GetLogger().Printf("[DEBUG] Writing buffer to socket: %v (client: %v).", c.conn, &c)
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
	utils.GetLogger().Println("[DEBUG] Finished writing buffer to socket.")

	// Waitgroup will be released when there's a response to the request.
	utils.GetLogger().Println("[DEBUG] Blocking until receive response to request.")
	// Time out if the message does not complete in time.
	// Adapted from: https://stackoverflow.com/questions/32840687/timeout-for-waitgroup-wait.
	select {
		case <- m.ch:
		case <- time.After(msgTimeout):
			// timed out
			return nil, errors.New("Timeout") 
	}
	utils.GetLogger().Println("[DEBUG] Received response to request.")

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
	utils.GetLogger().Println("[DEBUG] Extracted variables from message.")

	if err != io.EOF {
		return nil, err
	}
	err = nil

	if len(vars) > 0 {
		if e, ok := vars[len(vars)-1].(commError); ok {
			vars = vars[:len(vars)-1]
			if e.Error != "" {
				err = errors.New(e.Error)
			}
		}
	}

	return vars, err
}

func (c *client) HandleConnection() error {
	utils.GetLogger().Println("[INFO] Starting handling connection loop.")
	for {
		headerBuffer := make([]byte, 9)
		utils.GetLogger().Printf("[DEBUG] Reading header from socket: %v (client: %v).", c.conn, &c)
		_, err := c.conn.Read(headerBuffer)
		if err != nil {
			if err == io.EOF || !err.(net.Error).Temporary() {
				return err
			}
			continue
		}
		utils.GetLogger().Printf("[DEBUG] Read header into buffer: %v.", headerBuffer)

		response := false
		if headerBuffer[0] == 1 {
			response = true
		}
		messageID := binary.LittleEndian.Uint32(headerBuffer[1:5])
		messageLength := int(binary.LittleEndian.Uint32(headerBuffer[5:9]))
		utils.GetLogger().Printf("[DEBUG] Extracted from header isResponse: %v, messageID: %v, messageLength: %v.",
			response, messageID, messageLength)

		encryptedBuffer := make([]byte, messageLength)
		totalRead := 0

		utils.GetLogger().Println("[DEBUG] Reading contents from socket.")
		for totalRead < messageLength {
			read, err := c.conn.Read(encryptedBuffer[totalRead:])
			if err != nil {
				return err
			}
			totalRead += read
		}
		utils.GetLogger().Println("[DEBUG] Finished reading contents into buffer.")

		nonceSize := c.cipher.NonceSize()
		if len(encryptedBuffer) < nonceSize {
			return errors.New("ciphertext too short")
		}

		nonce, encryptedBuffer := encryptedBuffer[:nonceSize], encryptedBuffer[nonceSize:]
		buffer, err := c.cipher.Open(nil, nonce, encryptedBuffer, nil)
		if err != nil {
			return err
		}

		// Once the data is retrieved, process it in another thread so that we can continue receiving data.
		utils.GetLogger().Println("[DEBUG] Passing data processing to another thread.")
		go func() {
			err := c.processRequest(response, messageID, buffer)
			if err != nil {
				fmt.Println(err)
			}
		}()
	}
}

func (c *client) processRequest(response bool, messageID uint32, data []byte) error {
	utils.GetLogger().Printf("[DEBUG] Processing request isResponse: %v, messageID: %v.", response, messageID)
	if response {
		utils.GetLogger().Println("[DEBUG] Processing a request of response type.")
		c.messagesMutex.Lock()
		utils.GetLogger().Printf("[DEBUG] Removing message from messages map: %v.", c.messages)
		message, ok := c.messages[messageID]
		if ok {
			delete(c.messages, messageID)
			message.value = data
			message.ch <- struct{}{}
		}
		utils.GetLogger().Printf("[DEBUG] Updated message map: %v.", c.messages)
		c.messagesMutex.Unlock()

		return nil
	}

	utils.GetLogger().Println("[DEBUG] Processing request of non-response type.")
	// Extract the function name from the request.
	index := bytes.IndexByte(data, '\000')
	if index == -1 {
		fmt.Println(index, data)
		return c.respondWithError(messageID, errors.New("request incorrectly formed - could not extract function name"))
	}
	funcName := string(data[:index])
	utils.GetLogger().Printf("[DEBUG] Extracted function name: %v.", funcName)
	c.requestsMutex.RLock()
	utils.GetLogger().Printf("[DEBUG] Client's requests map: %v, request handlers: %v.", c.requests, c.requestsHandlers)
	request, ok := c.requests[funcName]
	utils.GetLogger().Printf("[DEBUG] Got request handler for function: %v.", request)
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
		return c.respondWithError(messageID, errors.New("function "+funcName+" is not registered."))
	}

	// Decode the buffer into variables.
	utils.GetLogger().Println("[DEBUG] Extracting variables from data.")
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
	utils.GetLogger().Println("[DEBUG] Finished extracting variables. Calling handler with variables.")
	returnVars := reflect.ValueOf(request).Call(vars)
	utils.GetLogger().Printf("[DEBUG] Return values of request handler with vars: %v.", returnVars)

	// If the last return argument is an error, change it to our 'error' type, so that we can recognise it later.
	if len(returnVars) > 0 && returnVars[len(returnVars)-1].Type().Implements(errorInterface) {
		if !returnVars[len(returnVars)-1].IsNil() {
			err := returnVars[len(returnVars)-1].Interface().(error)
			returnVars[len(returnVars)-1] = reflect.ValueOf(commError{err.Error()})
		} else {
			returnVars[len(returnVars)-1] = reflect.ValueOf(commError{""})
		}
	}

	// Encode the return variables and send them as reply.
	utils.GetLogger().Println("[DEBUG] Encoding returned values.")
	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)
	for i := range returnVars {
		d := returnVars[i].Interface()
		err := e.Encode(&d)
		if err != nil {
			return c.respondWithError(messageID, err)
		}
	}
	utils.GetLogger().Println("[DEBUG] Finished encoding return values.")

	// Encrypt the data using symmetric key.
	nonce := make([]byte, c.cipher.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return c.respondWithError(messageID, err)
	}

	encryptedBuffer := c.cipher.Seal(nonce, nonce, b.Bytes(), nil)

	buffer := make([]byte, 9+len(encryptedBuffer))
	buffer[0] = 1 // This is a response.
	binary.LittleEndian.PutUint32(buffer[1:5], messageID)
	binary.LittleEndian.PutUint32(buffer[5:9], uint32(len(encryptedBuffer)))
	copy(buffer[9:], encryptedBuffer)
	utils.GetLogger().Printf("[DEBUG] Created a response buffer (isResponse, messageID, length): %v.", buffer[:9])

	written := 0
	c.writeMutex.Lock()
	utils.GetLogger().Println("[DEBUG] Writing response to socket.")
	for written < len(buffer) {
		n, err := c.conn.Write(buffer[written:])
		if err != nil {
			return err
		}
		written += n
	}
	utils.GetLogger().Println("[DEBUG] Finished writing response to socket.")
	c.writeMutex.Unlock()
	return nil
}

func (c *client) respondWithError(messageID uint32, err error) error {
	utils.GetLogger().Println("[DEBUG] Responding with error", err)
	returnVars := []reflect.Value{reflect.ValueOf(commError{err.Error()})}

	// Encode the return error and send it as reply.
	utils.GetLogger().Println("[DEBUG] Encoding returned error.")
	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)
	for i := range returnVars {
		d := returnVars[i].Interface()
		err := e.Encode(&d)
		if err != nil {
			return err
		}
	}
	utils.GetLogger().Println("[DEBUG] Finished encoding return error.")

	// Encrypt the data using symmetric key.
	nonce := make([]byte, c.cipher.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}

	encryptedBuffer := c.cipher.Seal(nonce, nonce, b.Bytes(), nil)

	buffer := make([]byte, 9+len(encryptedBuffer))
	buffer[0] = 1 // This is a response.
	binary.LittleEndian.PutUint32(buffer[1:5], messageID)
	binary.LittleEndian.PutUint32(buffer[5:9], uint32(len(encryptedBuffer)))
	copy(buffer[9:], encryptedBuffer)
	utils.GetLogger().Printf("[DEBUG] Created a response buffer (isResponse, messageID, length): %v.", buffer[:9])

	written := 0
	c.writeMutex.Lock()
	utils.GetLogger().Println("[DEBUG] Writing response to socket.")
	for written < len(buffer) {
		n, err := c.conn.Write(buffer[written:])
		if err != nil {
			return err
		}
		written += n
	}
	utils.GetLogger().Println("[DEBUG] Finished writing response to socket.")
	c.writeMutex.Unlock()
	return nil
}
