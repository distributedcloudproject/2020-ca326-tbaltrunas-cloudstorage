package network

const (
	AuthMsg = "authenticate_me"
)

func createAuthRequestHandler(node *Node, network *Network) func(string) interface{} {
	r := request{
		network: network,
		node: node,
	}

	return func(message string) interface{} {
		switch message {
		case AuthMsg: return r.OnAuthenticateRequest
		}
		return nil
	}
}

func (n *Node) Authenticate() error {
	_, err := n.client.SendMessage(AuthMsg)
	return err
}

func (r request) OnAuthenticateRequest() {
	r.node.mutex.Lock()
	defer r.node.mutex.Unlock()

	if r.node.Authenticated {
		return
	}

	r.node.Authenticated = true
	r.node.client.AddRequestHandler(createRequestHandler(r.node, r.network))
}