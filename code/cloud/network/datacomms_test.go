package network

import (
	"cloud/datastore"
	"cloud/utils"
	"sort"
	"strconv"
	"testing"
	"time"
)

// TODO: fix this test after!
func TestNode_AddFileSaveChunk(t *testing.T) {
	numNodes := 4
	clouds, err := CreateTestClouds(numNodes)
	if err != nil {
		t.Fatal(err)
	}
	tmpStorageDirs, err := utils.GetTestDirs("cloud_test_node_data_", numNodes)
	if err != nil {
		t.Fatal(err)
	}
	defer utils.GetTestDirsCleanup(tmpStorageDirs)

	for i, cloud := range clouds {
		cloud.SetConfig(CloudConfig{
			FileStorageDir: tmpStorageDirs[i],
		})
	}

	t.Logf("Test clouds: %v.", clouds)
	t.Logf("Storage locations for clouds: %v.", tmpStorageDirs)
	cloud := clouds[0]
	t.Logf("Main cloud: %v.", cloud)
	t.Logf("MyNode on cloud: %v.", cloud.MyNode())
	t.Logf("Cloud with other nodes: %v.", cloud)
	t.Logf("Network: %v.", cloud.Network())
	nodes := cloud.Network().Nodes
	for i := range nodes {
		t.Logf("Node %d: %v.", i, nodes[i])
	}

	content := "hellothere i see you are a fan of bytes?" // 40 bytes
	contentBytes := []byte(content)
	tmpfile, err := utils.GetTestFile("cloud_test_file_*", contentBytes)
	if err != nil {
		t.Fatal(err)
	}
	defer utils.GetTestFileCleanup(tmpfile)

	chunkSize := 10 // will give 4 chunks
	file, err := datastore.NewFile(tmpfile, tmpfile.Name(), chunkSize)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("File: %v", file)

	n := nodes[0]
	t.Logf("Node: %v.", n)

	err = cloud.AddFile(file, "/test", tmpfile.Name())
	if err != nil {
		t.Error(err)
	}
	t.Logf("Network with added file: %v.", cloud.Network())

	//t.Log("Distributing chunks.")
	//// TODO: move to a function on its own
	//for i := 0; i < file.Chunks.NumChunks; i++ {
	//	t.Logf("Distributing chunk: %d (ID: %v), on node: %v", i, file.Chunks.Chunks[i].ID, cloud.Network().Nodes[i])
	//	err = cloud.GetCloudNode(cloud.Network().Nodes[i].ID).SaveChunk(file, i)
	//	if err != nil {
	//		t.Error(err)
	//	}
	//}

	// Check that we have a required ChunkNodes.
	t.Logf("Updated chunk-node locations: %v.", cloud.Network().ChunkNodes)
	chunks := file.Chunks.Chunks
	actualChunkNodes := cloud.Network().ChunkNodes
	allNodes := []string{cloud.Network().Nodes[0].ID, cloud.Network().Nodes[1].ID, cloud.Network().Nodes[2].ID, cloud.Network().Nodes[3].ID}
	expectedChunkNodes := ChunkNodes{
		chunks[0].ID: allNodes,
		chunks[1].ID: allNodes,
		chunks[2].ID: allNodes,
		chunks[3].ID: allNodes,
	}
	t.Logf("Expected ChunkNodes: %v.", expectedChunkNodes)
	t.Logf("Actial ChunkNodes: %v.", actualChunkNodes)
	// Note that DeepEqual has arguments against using it.
	// https://stackoverflow.com/a/45222521
	// An alternative struct comparison method may be needed in the future.
	// if !reflect.DeepEqual(cloud.Network.ChunkNodes, expectedChunkNodes) {
	// 	t.Error("ChunkNodes does not have the expected contents.")
	// }
	// FIXME: DeepEqual returns false. Need a better method.
	// Quick fix down below:
	if len(actualChunkNodes) != len(expectedChunkNodes) {
		t.Fatal("Actual and expected ChunkNodes do not match.", len(actualChunkNodes), len(expectedChunkNodes))
	}
	for k := range expectedChunkNodes {
		v := expectedChunkNodes[k]
		va := actualChunkNodes[k]
		sort.Strings(v)
		sort.Strings(va)
		for i, _ := range v {
			if v[i] != va[i] {
				t.Errorf("Element mismatch: %v, %v.", v[i], va[i])
			}
		}
	}
	for k := range actualChunkNodes {
		v := actualChunkNodes[k]
		va := expectedChunkNodes[k]
		sort.Strings(v)
		sort.Strings(va)
		for i, _ := range v {
			if v[i] != va[i] {
				t.Errorf("Element mismatch: %v, %v.", v[i], va[i])
			}
		}
	}

	// Check that all clouds have same ChunkNodes
	chunkLocations := cloud.Network().ChunkNodes
	t.Logf("ChunkNodes in main cloud representation: %v", chunkLocations)
	for _, c := range clouds {
		chunkLocationsOther := c.Network().ChunkNodes
		t.Logf("ChunkNodes in another cloud representation: %v.", chunkLocationsOther)
		for k := range chunkLocations {
			v := chunkLocations[k]
			va := chunkLocationsOther[k]
			sort.Strings(v)
			sort.Strings(va)
			for i, _ := range v {
				if v[i] != va[i] {
					t.Errorf("ChunkNodes mismatch %v %v", v[i], va[i])
				}
			}
		}
	}

	// Check that the storage benchmark state has been updated.
	spaceUsed := cloud.BenchmarkState().StorageSpaceUsed
	if spaceUsed != uint64(len(contentBytes)) {
		t.Errorf("Invalid benchmark state. Expected StorageSpaceUsed: %v. Got: %v.",
			len(contentBytes), spaceUsed)
	}
}

func TestNodeFileLock(t *testing.T) {
	key, _, err := createKey()
	if err != nil {
		t.Fatal(err)
	}

	cloud := SetupNetwork(Network{
		Name:        "My new network",
		Whitelist:   false,
		RequireAuth: true,
	}, Node{Name: "test"}, key)
	cloud.ListenOnPort(0)
	go cloud.Accept()

	var clouds []Cloud
	for i := 0; i < 4; i++ {
		key2, err := generateKey()
		if err != nil {
			t.Fatal(err)
		}

		n2, err := BootstrapToNetwork(cloud.MyNode().IP, Node{Name: "Node " + strconv.Itoa(i+1)}, key2, CloudConfig{})
		if err != nil {
			t.Fatal(err)
		}

		err = n2.ListenOnPort(0)
		if err != nil {
			t.Error(err)
		}
		go n2.Accept()
		clouds = append(clouds, n2)
	}
	time.Sleep(time.Millisecond * 100)

	if !cloud.LockFile("/file") {
		t.Fatal("could not lock /file")
	}

	if clouds[0].LockFile("/file") {
		t.Fatal("Node 1 acquired lock to /file; expected false")
	}
	clouds[0].UnlockFile("/file")
	if clouds[0].LockFile("/file") {
		t.Fatal("Node 1 unlocked and locked /file after; expected false")
	}
	cloud.UnlockFile("/file")
	if !clouds[0].LockFile("/file") {
		t.Fatal("Node 1 could not lock /file after it was unlocked")
	}
}

func TestNodeFileOperations(t *testing.T) {
	key, _, err := createKey()
	if err != nil {
		t.Fatal(err)
	}

	cloud := SetupNetwork(Network{
		Name:        "My new network",
		Whitelist:   false,
		RequireAuth: true,
	}, Node{Name: "test"}, key)
	cloud.ListenOnPort(0)
	go cloud.Accept()

	clouds := []Cloud{cloud}
	for i := 0; i < 4; i++ {
		key2, err := generateKey()
		if err != nil {
			t.Fatal(err)
		}

		n2, err := BootstrapToNetwork(cloud.MyNode().IP, Node{Name: "Node " + strconv.Itoa(i+1)}, key2, CloudConfig{})
		if err != nil {
			t.Fatal(err)
		}

		err = n2.ListenOnPort(0)
		if err != nil {
			t.Error(err)
		}
		go n2.Accept()
		clouds = append(clouds, n2)
	}
	time.Sleep(time.Millisecond * 100)

	// Create Directory.
	if err := cloud.CreateDirectory("/folder"); err != nil {
		t.Errorf("CreateDirectory(): %v", err)
	}
	if err := cloud.CreateDirectory("/folder2"); err != nil {
		t.Errorf("CreateDirectory(): %v", err)
	}
	if err := cloud.DeleteDirectory("/folder2"); err != nil {
		t.Errorf("CreateDirectory(): %v", err)
	}

	for _, c := range clouds {
		sub := c.Network().RootFolder.SubFolders
		if len(sub) != 1 || sub[0].Name != "folder" {
			t.Errorf("RootFolder directories: %v; expected: [folder]", sub)
			for _, s := range sub {
				t.Errorf(s.Name)
			}
		}
	}

	// Create File.
	if !cloud.LockFile("/folder/file") {
		t.Fatal("could not lock /file")
	}
	if !cloud.LockFile("/folder/file2") {
		t.Fatal("could not lock /file")
	}
	if !cloud.LockFile("/folder2/file") {
		t.Fatal("could not lock /file")
	}
	if err := cloud.AddFile(&datastore.File{
		ID:     "tempID",
		Name:   "file",
		Size:   0,
		Chunks: datastore.Chunks{},
	}, "/folder/file", ""); err != nil {
		t.Errorf("AddFile(): %v", err)
	}
	if err := cloud.AddFile(&datastore.File{
		ID:     "tempID",
		Name:   "file2",
		Size:   0,
		Chunks: datastore.Chunks{},
	}, "/folder/file2", ""); err != nil {
		t.Errorf("AddFile(): %v", err)
	}
	if err := cloud.DeleteFile("/folder/file2"); err != nil {
		t.Errorf("DeleteFile(): %v", err)
	}

	for _, c := range clouds {
		n := c.Network()
		nw, _ := n.GetFolder("/folder")
		if len(nw.Files.Files) != 1 || !nw.Files.ContainsName("file") {
			t.Errorf("Network files: %v; expected: [file]", nw.Files.Files)
		}
	}

	if err := cloud.MoveFile("/folder/file", "/folder2/file"); err != nil {
		t.Errorf("MoveFile(): %v", err)
	}
	for _, c := range clouds {
		n := c.Network()
		nw, _ := n.GetFolder("/folder2")
		if len(nw.Files.Files) != 1 || !nw.Files.ContainsName("file") {
			t.Errorf("Network files: %v; expected: [file]", nw.Files.Files)
		}
	}

	cloud.UnlockFile("/folder/file")
	cloud.UnlockFile("/folder/file2")
	cloud.UnlockFile("/folder2/file")
}
