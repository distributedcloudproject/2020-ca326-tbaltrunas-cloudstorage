package comm

import (
	"bytes"
	cipher2 "crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
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
	PublicKey() *rsa.PublicKey
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

	// Our private key, used for decryption.
	privateKey *rsa.PrivateKey

	// Their public key, used for encryption.
	publicKey *rsa.PublicKey

	// Master key used for symmetric encryption/decryption.
	masterKey []byte
	cipher cipher2.AEAD
}

// NewClientDial creates a new client by dialing the ip and creating a new socket connection.
func NewClientDial(address string, key *rsa.PrivateKey) (Client, error) {
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
	c.requests[message] = f
}

func (c *client) AddRequestHandler(handler RequestHandler) {
	c.requestsMutex.Lock()
	defer c.requestsMutex.Unlock()
	c.requestsHandlers = append(c.requestsHandlers, handler)
}

func (c *client) PublicKey() *rsa.PublicKey {
	return c.publicKey
}

// SendMessage sends a request with the msg and the data passed. Returns a list of arguments that were returned.
func (c *client) SendMessage(msg string, data ...interface{}) ([]interface{}, error) {
	id := atomic.AddUint32(&c.msgID, 1)

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

	// Place our message into the map, so that it can be used when receiving responses.
	m := &message{
		wg: sync.WaitGroup{},
	}
	c.messagesMutex.Lock()
	c.messages[id] = m
	c.messagesMutex.Unlock()

	m.wg.Add(1)

	// Write the buffer to the socket connection.
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

	// Waitgroup will be released when there's a response to the request.
	m.wg.Wait()

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

	if err != io.EOF {
		return nil, err
	}

	return vars, nil
}

func (c *client) HandleConnection() error {
	for {
		headerBuffer := make([]byte, 9)
		_, err := c.conn.Read(headerBuffer)
		if err != nil {
			if err == io.EOF || !err.(net.Error).Temporary() {
				return err
			}
			continue
		}

		response := false
		if headerBuffer[0] == 1 {
			response = true
		}
		messageID := binary.LittleEndian.Uint32(headerBuffer[1:5])
		messageLength := int(binary.LittleEndian.Uint32(headerBuffer[5:9]))

		encryptedBuffer := make([]byte, messageLength)
		totalRead := 0

		for totalRead < messageLength {
			read, err := c.conn.Read(encryptedBuffer[totalRead:])
			if err != nil {
				return err
			}
			totalRead += read
		}

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
		go func() {
			err := c.processRequest(response, messageID, buffer)
			if err != nil {
				fmt.Println(err)
			}
		}()
	}
}

func (c *client) processRequest(response bool, messageID uint32, data []byte) error {
	if response {
		c.messagesMutex.Lock()
		message, ok := c.messages[messageID]
		if ok {
			delete(c.messages, messageID)
			message.value = data
			message.wg.Done()
		}
		c.messagesMutex.Unlock()

		return nil
	}
	// Extract the function name from the request.
	index := bytes.IndexByte(data, '\000')
	if index == -1 {
		fmt.Println(index, data)
		return errors.New("request incorrectly formed - could not extract function name")
	}
	funcName := string(data[:index])
	c.requestsMutex.RLock()
	request, ok := c.requests[funcName]
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
	returnVars := reflect.ValueOf(request).Call(vars)

	// Encode the return variables and send them as reply.
	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)
	for i := range returnVars {
		d := returnVars[i].Interface()
		err := e.Encode(&d)
		if err != nil {
			return err
		}
	}

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

	written := 0
	c.writeMutex.Lock()
	for written < len(buffer) {
		n, err := c.conn.Write(buffer[written:])
		if err != nil {
			return err
		}
		written += n
	}
	c.writeMutex.Unlock()
	return nil
}
