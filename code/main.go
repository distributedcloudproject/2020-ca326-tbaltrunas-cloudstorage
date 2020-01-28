package main

import (
	"cloud/network"
	"flag"
	"fmt"
)


func main() {
	networkPtr := flag.String("network", "new", "Bootstrap IP of a node in an existing network or 'new' to create new network.")
	namePtr := flag.String("name", "", "Name of the node. Use for easy identification.")
	networkNamePtr := flag.String("network-name", "New Network", "The name of the network, if creating a new one")
	portPtr := flag.Int("port", 9000, "Port to listen on")

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
		err := n.Listen(*portPtr)
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
