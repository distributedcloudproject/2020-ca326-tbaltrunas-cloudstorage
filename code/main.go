package main

import (
	"cloud/network"
	"cloud/datastore"
	"cloud/utils"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"time"
	"path/filepath"
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
	verbosePtr := flag.Bool("verbose", false, "Print verbose information.")

	filePtr := flag.String("file", "", "A test file to save (back up) on the cloud.")
	fileStorageDirPtr := flag.String("file-storage-dir", filepath.Join("data", "user", "files"), 
									 "Directory where cloud files should be stored on the node.")

	logDirPtr := flag.String("log-dir", "", "The directory where logs should be written to.")
	logLevelPtr := flag.String("log-level", "WARN", fmt.Sprintf("The level of logging. One of: %v.", utils.LogLevels))

	flag.Parse()

	fmt.Println("Network:", *networkPtr)
	fmt.Println("Name:", *namePtr)
	fmt.Println("IP: ", *ipPtr+":"+strconv.Itoa(*portPtr))
	fmt.Println("Save File:", *saveFilePtr)
	fmt.Println("Test file to back up to the cloud:", *filePtr)
	fmt.Println("Directory for user file storage:", *fileStorageDirPtr)
	if *networkPtr == "new" {
		fmt.Println("Network Name:", *networkNamePtr)
	}

	fmt.Println("Log directory:", *logDirPtr)
	fmt.Println("Log level:", *logLevelPtr)

	if *logDirPtr != "" {
		err := os.MkdirAll(*logDirPtr, os.ModeDir)
		if err != nil {
			fmt.Println(err)
			return
		}
		t := time.Now()
		logFile := fmt.Sprintf("%v/%v.log", *logDirPtr, t.Format(time.RFC1123Z))
		f, err := os.Create(logFile)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer f.Close()

		if *logLevelPtr != "" {
			utils.NewLoggerFromWriterLevel(f, *logLevelPtr)
		} else {
			utils.NewLoggerFromWriter(f)
		}
	} else if *logLevelPtr != "" {
		utils.NewLoggerFromLevel(*logLevelPtr)
	}

	me := &network.Node{
		ID:   *idPtr,
		IP:   *ipPtr + ":" + strconv.Itoa(*portPtr),
		Name: *namePtr,
		FileStorageDir: *fileStorageDirPtr,
	}
	utils.GetLogger().Printf("[INFO] My node: %v.", me)

	var saveFunc func() io.Writer
	if *saveFilePtr != "" {
		utils.GetLogger().Println("[DEBUG] saveFilePtr is not empty. Creating saveFunc")
		saveFunc = func() io.Writer {
			f, _ := os.Create(*saveFilePtr)
			return f
		}
	}

	c := network.SetupNetwork(me, *networkNamePtr)
	c.SaveFunc = saveFunc
	utils.GetLogger().Printf("[INFO] Cloud: %v.", c)

	if *saveFilePtr != "" {
		utils.GetLogger().Println("[INFO] saveFilePtr is not empty. Loading from save file.")
		r, err := os.Open(*saveFilePtr)
		if err == nil {
			err := c.LoadNetwork(r)
			r.Close()
			if err != nil {
				fmt.Println(err)
				return
			}
		}
		utils.GetLogger().Printf("[INFO] Loaded cloud: %v.", c)
	}

	if *networkPtr != "new" {
		utils.GetLogger().Println("[INFO] Boostrapping to an existing network.")
		// TODO: Verify ip is a valid ip.
		ip := *networkPtr
		n, err := network.BootstrapToNetwork(ip, me)
		n.SaveFunc = saveFunc
		if err != nil {
			fmt.Println(err)
			return
		}
		c = n
		utils.GetLogger().Printf("[INFO] Bootstrapped cloud: %v.", c)
	}

	if *fancyDisplayPtr {
		utils.GetLogger().Println("[INFO] Initialising fancy display.")
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
				if *verbosePtr {
					fmt.Printf("DataStore: %v | FileChunkLocations: %v\n", 
							   c.Network.DataStore, c.Network.FileChunkLocations)
				}
				fmt.Printf("Name, ID, Online[, Node]:\n")
				for _, n := range c.Network.Nodes {
					row := fmt.Sprintf("|%-20v|%-20v|%-8v|", n.Name, n.ID, n.Online())
					if *verbosePtr {
						row += fmt.Sprintf("\t%v|", n)
					}
					fmt.Printf("%v\n", row)
				}
			}
		}(c)
	}

	if *filePtr != "" && *fileStorageDirPtr != "" {
		fmt.Println("Storing user file: ", *filePtr)
		r, err := os.Open(*filePtr)
		defer r.Close()
		if err != nil {
			fmt.Println(err)
			return
		}
		file, err := datastore.NewFile(r, *filePtr, 5)
		if err != nil {
			fmt.Println(err)
			return
		}

		for i := range c.Network.Nodes {
			fmt.Println(c.Network.Nodes[i])
			err = c.Network.Nodes[i].AddFile(file)
			if err != nil {
				fmt.Println(err)
				return
			}
		}

		// err = c.MyNode.AddFile(file)
		// if err != nil {
		// 	fmt.Println(err)
		// 	return
		// }

		// i := 0
		// err = c.MyNode.SaveChunk(file, i)
		// if err != nil {
		// 	fmt.Println(err)
		// 	return
		// }

		// for i := range c.Network.Nodes {
		// 	fmt.Println(c.Network.Nodes[i])
		// 	err = c.Network.Nodes[i].SaveChunk(file, i)
		// 	if err != nil {
		// 		fmt.Println(err)
		// 		return
		// 	}
		// }
	}

	utils.GetLogger().Println("[INFO] Initialising listening.")
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
