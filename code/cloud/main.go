package main

import (
	"bufio"
	"cloud/datastore"
	"cloud/network"
	"cloud/utils"
	"cloud/webapp"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	_ "github.com/joho/godotenv/autoload" // automatically load environment variables from .env file
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
	fileStorageDirPtr := flag.String("file-storage-dir", "", "Directory where cloud files should be stored on the node.")
	fileStorageCapacityPtr := flag.Int64("file-storage-capacity", 0, "Storage space in bytes allocated for file storage.")
	fileChunkSizePtr := flag.Int("file-chunk-size", 10*1e+7, "Chunk size in bytes used for file splitting (default 10 megabytes)")

	logDirPtr := flag.String("log-dir", "", "The directory where logs should be written to.")
	logLevelPtr := flag.String("log-level", "WARN", fmt.Sprintf("The level of logging. One of: %v.", utils.LogLevels))

	webBackendPtr := flag.Bool("web-backend", false, "Enable web backend on this node.")
	webBackendPortPtr := flag.Int("web-backend-port", 9443, "Port to listen on for web API requests.")

	flag.Parse()

	fmt.Println("Network:", *networkPtr)
	fmt.Println("Name:", *namePtr)
	fmt.Println("IP:", *ipPtr+":"+strconv.Itoa(*portPtr))
	fmt.Println("Save File:", *saveFilePtr)

	fmt.Println("Secure:", *networkSecurePtr)
	fmt.Println("Whitelist:", *networkWhitelistPtr)

	fmt.Println("Test file to back up to the cloud:", *filePtr)
	fmt.Println("Directory for user file storage:", *fileStorageDirPtr)

	fmt.Println("Log directory:", *logDirPtr)
	fmt.Println("Log level:", *logLevelPtr)

	fmt.Println("Web backend: ", *webBackendPtr)
	fmt.Println("Web backend port: ", *webBackendPortPtr)

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

	me := network.Node{
		ID:        id,
		IP:        *ipPtr + ":" + strconv.Itoa(*portPtr),
		Name:      *namePtr,
		PublicKey: key.PublicKey,
	}
	utils.GetLogger().Printf("[INFO] My node: %v.", me)

	var c network.Cloud
	if *networkPtr == "new" {
		c = network.SetupNetwork(network.Network{
			Name:        *networkNamePtr,
			Whitelist:   *networkWhitelistPtr,
			RequireAuth: *networkSecurePtr,
		}, me, key)
	} else {
		utils.GetLogger().Println("[INFO] Bootstrapping to an existing network.")
		// TODO: Verify ip is a valid ip.
		ip := *networkPtr
		n, err := network.BootstrapToNetwork(ip, me, key, network.CloudConfig{FileStorageDir: *fileStorageDirPtr})
		if err != nil {
			fmt.Println(err)
			return
		}
		c = n
		utils.GetLogger().Printf("[INFO] Bootstrapped cloud: %v.", c)
	}

	c.SetConfig(network.CloudConfig{
		FileStorageDir:      *fileStorageDirPtr,
		FileStorageCapacity: *fileStorageCapacityPtr,
		FileChunkSize:       *fileChunkSizePtr,
	})

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

	if *fancyDisplayPtr {
		utils.GetLogger().Println("[INFO] Initialising fancy display.")
		go func(c network.Cloud) {
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

				network := c.Network()
				fmt.Printf("Network: %s | Nodes: %d | Online: %d\n", network.Name, len(network.Nodes), c.OnlineNodesNum())
				fmt.Printf("Name, ID, Online:\n")
				for _, n := range network.Nodes {
					fmt.Printf("|%-20v|%-20v|%8v|\n", n.Name, n.ID, c.IsNodeOnline(n.ID))
				}
				if *verbosePtr {
					fmt.Printf("Root folder: %v\n", c.Network().RootFolder)
					fmt.Printf("ChunkNodes: %v\n",
						network.ChunkNodes)
					fmt.Printf("My node: %v.", c.MyNode())
				}
			}
		}(c)
	}

	if *filePtr != "" && *fileStorageDirPtr != "" {
		fmt.Println("Storing user file: ", *filePtr)
		time.Sleep(time.Millisecond * 100)
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

		err = c.AddFile(file, "/"+file.Name, *filePtr)
		if err != nil {
			fmt.Println(err)
			return
		}

		numReplicas := -1
		antiAffinity := true
		err = c.Distribute("/"+file.Name, *file, numReplicas, antiAffinity)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	if *webBackendPtr {
		fmt.Println("Web backend enabled. Listening on: ", *webBackendPortPtr)
		wapp := webapp.NewWebApp(c)
		go wapp.Serve(*webBackendPortPtr)
	}

	utils.GetLogger().Println("[INFO] Initialising listening.")
	err = c.ListenOnPort(*portPtr)
	if err != nil {
		fmt.Println(err)
		return
	}
	go c.Accept()
	for {
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		cmd := strings.Split(text, " ")
		if cmd[0] == "syncfolder" {
			if len(cmd) == 3 {
				fmt.Println("Syncing cloud:", cmd[1], "to local folder:", cmd[2])
				err := c.SyncFolder(cmd[1], cmd[2])
				if err != nil {
					fmt.Println("SyncFolder err: ", err)
				}
			} else {
				fmt.Println("Usage: syncfolder <cloud folder> <local folder>")
			}
		}
		if cmd[0] == "dir" {
			if len(cmd) == 1 {
				fmt.Println("sub-commands available: [list, create]")
				continue
			}
			if cmd[1] == "create" {
				if len(cmd) != 3 {
					fmt.Println("Usage: dir create <cloud dir>")
					continue
				}
				err := c.CreateDirectory(cmd[2])
				if err != nil {
					fmt.Println("Directory Create error:", err)
				} else {
					fmt.Println("Directory Created: ", cmd[2])
				}
			}
			if cmd[1] == "list" {
				if len(cmd) != 3 {
					fmt.Println("Usage: dir list <cloud dir>")
					continue
				}
				nw, err := c.GetFolder(cmd[2])
				if err != nil {
					fmt.Println("Directory list error:", err)
				} else {
					fmt.Println("Directory list: ", cmd[2])
					for _, sub := range nw.SubFolders {
						fmt.Println(sub.Name)
					}
					fmt.Println("")
					for _, f := range nw.Files.Files {
						fmt.Println(f.Name)
					}
				}
			}
		}
		if cmd[0] == "whitelist" {
			if len(cmd) == 1 {
				fmt.Println("sub-commands available: [list, add]")
				continue
			}
			if cmd[1] == "add" {
				if len(cmd) != 3 {
					fmt.Println("Usage: whitelist add <ID>")
					continue
				}
				err := c.AddToWhitelist(cmd[2])
				if err != nil {
					fmt.Println("Whitelist Add error:", err)
				} else {
					fmt.Println("Whitelist added: ", cmd[2])
				}
			}
			if cmd[1] == "list" {
				fmt.Println("Whitelist:", c.Whitelist())
			}
		}
	}
}
