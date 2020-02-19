package main

import (
	"bufio"
	"cloud/datastore"
	"cloud/distribution"
	"cloud/network"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"cloud/utils"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)


func readKey(file string) (*rsa.PrivateKey, error) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	bb, _ := pem.Decode(b)
	if bb.Type != "RSA PRIVATE KEY" {
		return nil, errors.New("invalid type " + bb.Type + " want: RSA PRIVATE KEY")
	}

	key, err := x509.ParsePKCS1PrivateKey(bb.Bytes)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func main() {
	networkPtr := flag.String("network", "new", "Bootstrap IP of a node in an existing network or 'new' to create new network.")
	networkNamePtr := flag.String("network-name", "New Network", "The name of the network, if creating a new one.")
	networkSecurePtr := flag.Bool("secure", true, "Enable authentication for the network.")
	saveFilePtr := flag.String("save-file", "", "File to save network state and resume network state from.")
	networkWhitelistPtr := flag.Bool("whitelist", true, "Enable whitelist for cloud. Node IDs will need to be whitelisted before joining the network.")
	networkWhitelistFilePtr := flag.String("whitelist-file", "", "Load node IDs from file into the whitelist. 1 per line.")

	namePtr := flag.String("name", "", "Name of the node. Use for easy identification.")
	privateKeyPtr := flag.String("key", "", "Path to private key.")
	ipPtr := flag.String("ip", "", "Remote IP to override source IP address when connecting to local nodes.")
	portPtr := flag.Int("port", 9000, "Port to listen on.")

	fancyDisplayPtr := flag.Bool("fancy-display", false, "Display node information in a fancy-way.")
	verbosePtr := flag.Bool("verbose", false, "Print verbose information.")

	filePtr := flag.String("file", "", "A test file to save (back up) on the cloud.")
	fileStorageDirPtr := flag.String("file-storage-dir", "", 
									 "Directory where cloud files should be stored on the node.")

	logDirPtr := flag.String("log-dir", "", "The directory where logs should be written to.")
	logLevelPtr := flag.String("log-level", "WARN", fmt.Sprintf("The level of logging. One of: %v.", utils.LogLevels))

	flag.Parse()

	fmt.Println("Network:", *networkPtr)
	fmt.Println("Name:", *namePtr)
	fmt.Println("IP:", *ipPtr+":"+strconv.Itoa(*portPtr))
	fmt.Println("Save File:", *saveFilePtr)

	fmt.Println("Secure:", *networkSecurePtr)
	fmt.Println("Whitelist:", *networkWhitelistPtr)

	fmt.Println("Test file to back up to the cloud:", *filePtr)
	fmt.Println("Directory for user file storage:", *fileStorageDirPtr)

	if *networkPtr == "new" {
		fmt.Println("Network Name:", *networkNamePtr)
	}

	// Read the key.
	key, err := readKey(*privateKeyPtr)
	if err != nil {
		fmt.Println("Error while parsing key:", err)
		return
	}
	id, err := network.PublicKeyToID(&key.PublicKey)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("ID:", id)
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
		ID:   id,
		IP:   *ipPtr + ":" + strconv.Itoa(*portPtr),
		Name: *namePtr,
		PublicKey: key.PublicKey,
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

	c := network.SetupNetwork(me, *networkNamePtr, key)
	c.SaveFunc = saveFunc
	c.Network.RequireAuth = *networkSecurePtr
	c.Network.Whitelist = *networkWhitelistPtr

	if *networkWhitelistFilePtr != "" {
		r, err := os.Open(*networkWhitelistFilePtr)
		if err == nil {
			reader := bufio.NewReader(r)
			for {
				line, err := reader.ReadString('\n')
				c.AddToWhitelist(strings.TrimSpace(line))
				if err != nil {
					break
				}
			}
			r.Close()
		}
	}

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
		n, err := network.BootstrapToNetwork(ip, me, key)
		if err != nil {
			fmt.Println(err)
			return
		}
		n.SaveFunc = saveFunc
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
					fmt.Printf("DataStore: %v | ChunkNodes: %v\n", 
							   c.Network.DataStore, c.Network.ChunkNodes)
					fmt.Printf("My node: %v.", c.MyNode)
				}
				fmt.Printf("Name, ID, Online[, Node]:\n")
				for _, n := range c.Network.Nodes {
					fmt.Printf("|%-20v|%-20v|%-8v|\n", n.Name, n.ID, n.Online())
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

		err = c.MyNode.AddFile(file)
		if err != nil {
			fmt.Println(err)
			return
		}
		err = distribution.Distribute(file, c)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	utils.GetLogger().Println("[INFO] Initialising listening.")
	err = c.Listen(*portPtr)
	if err != nil {
		fmt.Println(err)
		return
	}
	c.AcceptListener()
}

