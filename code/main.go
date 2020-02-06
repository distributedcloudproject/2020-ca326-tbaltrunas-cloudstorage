package main

import (
	"cloud/network"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"time"
)


func main() {
	networkPtr := flag.String("network", "new", "Bootstrap IP of a node in an existing network or 'new' to create new network.")
	networkNamePtr := flag.String("network-name", "New Network", "The name of the network, if creating a new one")
	saveFilePtr := flag.String("save-file", "Save File", "File to save network state and resume network state from.")

	namePtr := flag.String("name", "", "Name of the node. Use for easy identification.")
	idPtr := flag.String("id", "Node ID", "Temporary")
	ipPtr := flag.String("ip", "", "Remote IP to override source IP address when connecting to local nodes")
	portPtr := flag.Int("port", 9000, "Port to listen on")

	fancyDisplayPtr := flag.Bool("fancy-display", false, "Display node information in a fancy-way.")

	flag.Parse()

	fmt.Println("Network:", *networkPtr)
	fmt.Println("Name:", *namePtr)
	fmt.Println("IP: ", *ipPtr+":"+strconv.Itoa(*portPtr))
	fmt.Println("Save File:", *saveFilePtr)
	if *networkPtr == "new" {
		fmt.Println("Network Name:", *networkNamePtr)
	}

	me := &network.Node{
		ID:   *idPtr,
		IP:   *ipPtr + ":" + strconv.Itoa(*portPtr),
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
			Name:  *networkNamePtr,
			Nodes: []*network.Node{me},
		},
		MyNode:   me,
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

	if *fancyDisplayPtr {
		go func(c *network.Cloud) {
			for {
				time.Sleep(time.Second * 1)
				switch runtime.GOOS {
				case "linux":
					{
						cmd := exec.Command("clear")
						cmd.Stdout = os.Stdout
						cmd.Run()
					}
				case "windows":
					{
						cmd := exec.Command("cmd", "/c", "cls")
						cmd.Stdout = os.Stdout
						cmd.Run()
					}
				}
				fmt.Printf("Network: %s | Nodes: %d | Online: %d\n", c.Network.Name, len(c.Network.Nodes), c.OnlineNodesNum())
				for _, n := range c.Network.Nodes {
					fmt.Printf("|%-20v|%-20v|%8v|\n", n.Name, n.ID, n.Online())
				}
			}
		}(c)
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
