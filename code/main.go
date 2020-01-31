package main

import (
	"cloud/network"
	"flag"
	"fmt"
	"strconv"
)

type File struct {
}

func main() {
	networkPtr := flag.String("network", "new", "Bootstrap IP of a node in an existing network or 'new' to create new network.")
	networkNamePtr := flag.String("network-name", "New Network", "The name of the network, if creating a new one")

	namePtr := flag.String("name", "", "Name of the node. Use for easy identification.")
	ipPtr := flag.String("ip", "", "Remote IP to override source IP address when connecting to local nodes")
	portPtr := flag.Int("port", 9000, "Port to listen on")

	flag.Parse()

	fmt.Println("Network:", *networkPtr)
	fmt.Println("Name:", *namePtr)
	fmt.Println("IP: ", *ipPtr + ":" + strconv.Itoa(*portPtr))
	if *networkPtr == "new" {
		fmt.Println("Network Name:", *networkNamePtr)
	}

	me := &network.Node{
		ID: "0",
		IP: *ipPtr + ":" + strconv.Itoa(*portPtr),
		Name: *namePtr,
	}

	if *networkPtr == "new" {
		n := &network.Cloud{
			Network: network.Network{
				Name: *networkNamePtr,
				Nodes: []*network.Node{me},
			},
			MyNode: me,
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
		n, err := network.BootstrapToNetwork(ip, me)
		if err != nil {
			fmt.Println(err)
			return
		}
		err = n.Listen(*portPtr)
		if err != nil {
			fmt.Println(err)
			return
		}
		n.AcceptListener()
	}
}

func ExploreNode(ip string) {
	//
}
