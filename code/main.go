package main

import (
	"cloud/network"
	"flag"
	"fmt"
)

type File struct {
}

func main() {
	networkPtr := flag.String("network", "new", "Bootstrap IP of a node in an existing network or 'new' to create new network.")
	namePtr := flag.String("name", "", "Name of the node. Use for easy identification.")
	networkNamePtr := flag.String("network-name", "New Network", "The name of the network, if creating a new one")

	flag.Parse()

	fmt.Println("Network:", *networkPtr)
	fmt.Println("Name:", *namePtr)
	if *networkPtr == "new" {
		fmt.Println("Network Name:", *networkNamePtr)
	}

	if *networkPtr == "new" {
		n := &network.Network{
			Name: *networkNamePtr,
		}
		err := n.Listen(9000)
		if err != nil {
			fmt.Println(err)
			return
		}
		n.AcceptListener()
	} else {
		// TODO: Verify ip is a valid ip.
		ip := *networkPtr
		network.BootstrapToNetwork(ip)
	}
}

func ExploreNode(ip string) {
	//
}
