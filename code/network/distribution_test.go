package network

import (
	"cloud/datastore"
	"cloud/utils"
	"reflect"
	"sort"
	"testing"
)

// map from node indices to a slice of chunk indices (sequence number)
// node indices are in terms of the first cloud.
type testCaseDistribution map[int][]int

// TODO: might want to split multiple tests into multiple functions, to be able to rerun only 1 test case when needed

func TestDistribution(t *testing.T) {
	numNodes := 2
	chunkSize := 5

	testCases := []struct {
		Contents 			string
		StorageCapacities 	[]int64
		NumReplicas 		int
		AntiAffinity 		bool

		Distribution 		testCaseDistribution
	}{
		// Two nodes with equal storage.
		// Both get a chunk.
		{
			Contents: "hellothere", // 2 chunks
			StorageCapacities: []int64{100, 100},
			NumReplicas: 0,
			AntiAffinity: true,
			Distribution: testCaseDistribution{
				0: []int{0},
				1: []int{1},				
			},
		},
		// One large node, one small.
		// Large node gets both chunks.
		{
			Contents: "hellothere", // 2 chunks
			StorageCapacities: []int64{500, 100},
			NumReplicas: 0,
			AntiAffinity: true,
			Distribution: testCaseDistribution{
				0: []int{0, 1},				
			},
		},
		// One large node, one small.
		// Both nodes get replicas of one chunk (anti-affinity rule).
		{
			Contents: "hello", // 1 chunk
			StorageCapacities: []int64{500, 100},
			NumReplicas: 1, // 2 chunks
			AntiAffinity: true,
			Distribution: testCaseDistribution{
				0: []int{0},
				1: []int{0},
			},
		},
		// Special replication case.
		// All nodes get all chunks.
		{
			Contents: "helloworld", // 2 chunks
			StorageCapacities: []int64{200, 100},
			NumReplicas: -1,
			AntiAffinity: true,
			Distribution: testCaseDistribution{
				0: []int{0, 1},
				1: []int{0, 1},
			},
		},
	}

	for i, testCase := range testCases {
		t.Logf("Case: %d.", i)
		// FIXME: optimize by not creating a new cloud each time, i.e. reset ChunkNodes
		clouds, err := CreateTestClouds(numNodes)
		if err != nil {
			t.Fatal(err)
		}
		cloud := clouds[0]
		for i, n := range 	cloud.Network().Nodes {
			t.Logf("Node %d: %v.", i + 1, n.ID)
		}

		tmpfile, err := utils.GetTestFile("cloud_test_file_*", []byte(testCase.Contents))
		if err != nil {
			t.Fatal(err)
		}
		defer utils.GetTestFileCleanup(tmpfile)

		file, err := datastore.NewFile(tmpfile, tmpfile.Name(), chunkSize)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("File: %v", file)

		tmpStorageDirs, err := utils.GetTestDirs("cloud_test_node_data_", numNodes)
		if err != nil {
			t.Fatal(err)
		}
		defer utils.GetTestDirsCleanup(tmpStorageDirs)
		t.Logf("Temporary directories for nodes: %v.", tmpStorageDirs)

		for i, cloud := range clouds {
			cloud.SetConfig(CloudConfig{
				FileStorageDir: tmpStorageDirs[i],
				FileStorageCapacity: testCase.StorageCapacities[i],
			})
		}

		err = cloud.Distribute(*file, testCase.NumReplicas, testCase.AntiAffinity)
		if err != nil {
			t.Error(err)
		}

		// Create a ChunkNodes object from the test case map
		expectedChunkNodes := make(ChunkNodes)
		for nodeIndex, chunkIndices := range testCase.Distribution {
			nodeID := cloud.Network().Nodes[nodeIndex].ID
			for _, chunkIndex := range chunkIndices {
				chunkID := file.Chunks.Chunks[chunkIndex].ID
				nodeIDs, ok := expectedChunkNodes[chunkID]
				if ok {
					nodeIDs = append(nodeIDs, nodeID)
				} else {
					nodeIDs = []string{nodeID}
				}
				expectedChunkNodes[chunkID] = nodeIDs
			}
		}

		// sort for working comparison
		for _, nodeIDs := range expectedChunkNodes {
			sort.Strings(nodeIDs)
		}
		for _, nodeIDs := range cloud.Network().ChunkNodes {
			sort.Strings(nodeIDs)
		}
		
		if !reflect.DeepEqual(expectedChunkNodes, cloud.Network().ChunkNodes) {
			t.Errorf("case(%d).Distribution got ChunkNodes %v; want %v", i, cloud.Network().ChunkNodes, expectedChunkNodes)
		}
	}
}

func TestDistributionNoStore(t *testing.T) {
	clouds, err := CreateTestClouds(1)
	if err != nil {
		t.Fatal(err)
	}
	cloud := clouds[0]
	cloud.SetConfig(CloudConfig{
		FileStorageCapacity: -1,
	})

	tmpfile, err := utils.GetTestFile("cloud_test_file_*", []byte("hello"))
	if err != nil {
		t.Fatal(err)
	}
	defer utils.GetTestFileCleanup(tmpfile)
	file, err := datastore.NewFile(tmpfile, tmpfile.Name(), 5)
	if err != nil {
		t.Fatal(err)
	}

	err = cloud.Distribute(*file, 0, true)
	if err == nil {
		t.Errorf("Expected error. Got: %v.", err)
	} else if err.Error() != "No nodes available" {
		t.Errorf("Expected error message: No node available. Got: %s.", err.Error())
	}
}

func TestDistributionError(t *testing.T) {
	clouds, err := CreateTestClouds(1)
	if err != nil {
		t.Fatal(err)
	}
	cloud := clouds[0]
	var file datastore.File
	err = cloud.Distribute(file, -2, true)
	if err == nil {
		t.Fatalf("Got err: %v. Expected non-nil err.", err)
	}
}

// TODO: tests that measure optimal network balancing
// For now test in system test (real node tests).