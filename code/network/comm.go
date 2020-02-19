package network

type Response struct {
	Returns []interface{}
	Error   error
	Node    *cloudNode
}

func (c *cloud) SendMessageAllOthers(msg string, any ...interface{}) []Response {
	responseChannel := make(chan Response)
	c.NodesMutex.RLock()
	sent := 0
	myID := c.MyNode().ID

	for _, n := range c.Nodes {
		if n.ID == myID {
			continue
		}
		sent++
		go func(node *cloudNode, res chan Response) {
			ret, err := n.client.SendMessage(msg, any...)
			res <- Response{
				Returns: ret,
				Error:   err,
				Node:    node,
			}
		}(n, responseChannel)
	}
	c.NodesMutex.RUnlock()

	responses := make([]Response, sent)
	for i := 0; i < sent; i++ {
		res := <-responseChannel
		responses[i] = res
	}
	return responses
}

func (c *cloud) SendMessageAll(msg string, any ...interface{}) []Response {
	responseChannel := make(chan Response)
	c.NodesMutex.RLock()
	sent := 0
	for _, n := range c.Nodes {
		sent++
		go func(node *cloudNode, res chan Response) {
			ret, err := n.client.SendMessage(msg, any...)
			res <- Response{
				Returns: ret,
				Error:   err,
				Node:    node,
			}
		}(n, responseChannel)
	}
	c.NodesMutex.RUnlock()

	responses := make([]Response, sent)
	for i := 0; i < sent; i++ {
		res := <-responseChannel
		responses[i] = res
	}
	return responses
}

func (c *cloud) SendMessageToMe(msg string, any ...interface{}) ([]interface{}, error) {
	myID := c.MyNode().ID
	node := c.GetCloudNode(myID)
	return node.client.SendMessage(msg, any...)
}
