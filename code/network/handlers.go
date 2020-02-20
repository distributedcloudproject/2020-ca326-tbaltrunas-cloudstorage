package network

var handlers []func(n *cloudNode, c *cloud) func(string) interface{}

func (c *cloud) addRequestHandlers(n *cloudNode) {
	for h := range handlers {
		n.client.AddRequestHandler(handlers[h](n, c))
	}
}
