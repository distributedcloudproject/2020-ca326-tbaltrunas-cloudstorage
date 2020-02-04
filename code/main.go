package main

import (
	"cloud/network"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
)

type File struct {
}

func main() {
	networkPtr := flag.String("network", "new", "Bootstrap IP of a node in an existing network or 'new' to create new network.")
	networkNamePtr := flag.String("network-name", "New Network", "The name of the network, if creating a new one")
	saveFilePtr := flag.String("save-file", "Save File", "File to save network state and resume network state from.")

	namePtr := flag.String("name", "", "Name of the node. Use for easy identification.")
	idPtr := flag.String("id", "Node ID", "Temporary")
	ipPtr := flag.String("ip", "", "Remote IP to override source IP address when connecting to local nodes")
	portPtr := flag.Int("port", 9000, "Port to listen on")

	flag.Parse()

	fmt.Println("Network:", *networkPtr)
	fmt.Println("Name:", *namePtr)
	fmt.Println("IP: ", *ipPtr + ":" + strconv.Itoa(*portPtr))
	fmt.Println("Save File:", *saveFilePtr)
	if *networkPtr == "new" {
		fmt.Println("Network Name:", *networkNamePtr)
	}

	me := &network.Node{
		ID: *idPtr,
		IP: *ipPtr + ":" + strconv.Itoa(*portPtr),
		Name: *namePtr,
	}

	var saveFunc func() io.Writer
	if *saveFilePtr != "" {
		saveFunc = func() io.Writer {
			f, _ := os.Create(*saveFilePtr)
			return f
		}
	}

	c := &network.Cloud{
		Network: network.Network{
			Name: *networkNamePtr,
			Nodes: []*network.Node{me},
		},
		MyNode: me,
		SaveFunc: saveFunc,
	}

	if *saveFilePtr != "" {
		r, err := os.Open(*saveFilePtr)
		if err == nil {
			err := c.LoadNetwork(r)
			r.Close()
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}

	if *networkPtr != "new" {
		// TODO: Verify ip is a valid ip.
		ip := *networkPtr
		n, err := network.BootstrapToNetwork(ip, me)
		n.SaveFunc = saveFunc
		if err != nil {
			fmt.Println(err)
			return
		}
		c = n
	}

	err := c.Listen(*portPtr)
	if err != nil {
		fmt.Println(err)
		return
	}
	c.AcceptListener()
}

func ExploreNode(ip string) {
	//
}
