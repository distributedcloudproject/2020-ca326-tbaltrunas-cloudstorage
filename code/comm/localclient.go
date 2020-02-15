package comm

import (
	"errors"
	"reflect"
	"sync"
)

/*
type Client interface {
	RegisterRequest(message string, f interface{})
	AddRequestHandler(handler RequestHandler)

	Address() string

	SendMessage(msg string, data ...interface{}) ([]interface{}, error)
	HandleConnection() error
	Close() error
}
 */

type localClient struct {
	requests map[string]interface{}
	requestsHandlers []RequestHandler
	requestsMutex sync.RWMutex
}

func NewLocalClient() Client {
	return &localClient{
		requests: make(map[string]interface{}),
	}
}

func (c *localClient) Address() string {
	return "localhost"
}

func (c *localClient) HandleConnection() error {
	return nil
}

func (c *localClient) Close() error {
	return nil
}

func (c *localClient) RegisterRequest(message string, f interface{}) {
	c.requestsMutex.Lock()
	defer c.requestsMutex.Unlock()
	c.requests[message] = f
}

func (c *localClient) AddRequestHandler(handler RequestHandler) {
	c.requestsMutex.Lock()
	defer c.requestsMutex.Unlock()
	c.requestsHandlers = append(c.requestsHandlers, handler)
}

// SendMessage sends a request with the msg and the data passed. Returns a list of arguments that were returned.
func (c *localClient) SendMessage(msg string, data ...interface{}) ([]interface{}, error) {
	c.requestsMutex.RLock()
	request, ok := c.requests[msg]
	if !ok {
		for i := 0; i < len(c.requestsHandlers); i++ {
			request = c.requestsHandlers[i](msg)
			if request != nil {
				ok = true
				break
			}
		}
	}
	c.requestsMutex.RUnlock()

	if !ok {
		return nil, errors.New("function " + msg + " is not registered.")
	}

	// Decode the buffer into variables.
	vars := make([]reflect.Value, 0)
	for i := range data {
		vars = append(vars, reflect.ValueOf(data[i]))
	}
	returnVars := reflect.ValueOf(request).Call(vars)

	returnVarsInterface := make([]interface{}, len(returnVars))
	for i := range returnVars {
		returnVarsInterface[i] = returnVars[i].Interface()
	}

	return returnVarsInterface, nil
}