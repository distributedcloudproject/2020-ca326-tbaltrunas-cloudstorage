package screens

import (
	"cloud/network"
	"fmt"
)

func RegisterEvents(c network.Cloud) {
	c.Events().NodeAdded = func(n network.Node) {
		nodesEventNodeAdded(n)
	}
	c.Events().NodeConnected = func(ID string) {
		fmt.Println("Node connected: ", ID)
		nodesEventNodeConnected(ID)
		homeEventNodeConnected(ID)
	}
	c.Events().NodeDisconnected = func(ID string) {
		nodesEventNodeDisconnected(ID)
		homeEventNodeDisconnected(ID)
	}
	c.Events().NodeUpdated = func(n network.Node) {
		nodesEventNodeUpdated(n)
	}
}
