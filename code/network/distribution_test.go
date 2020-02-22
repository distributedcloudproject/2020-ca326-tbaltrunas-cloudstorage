package network

import (
	"cloud/datastore"
	"cloud/utils"
	"reflect"
	"testing"
)

// map from chunk idx to node idx's
type testCaseDistribution map[int][]int

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
		{
			Contents: "hellothere", // 2 chunks
			StorageCapacities: []int64{500, 100},
			NumReplicas: 0,
			AntiAffinity: true,
			Distribution: testCaseDistribution{
				0: []int{0},
				1: []int{0},				
			},
		},
		// {
		// 	Contents: "hello", // 1 chunk
		// 	StorageCapacities: []int64{500, 100},
		// 	NumReplicas: 1, // 2 chunks
		// 	AntiAffinity: true,
		// 	Distribution: testCaseDistribution{
		// 		0: []int{0, 1},				
		// 	},
		// },
	}

	// Distribute file
	for i, testCase := range testCases {
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

		err = Distribute(file, cloud, testCase.NumReplicas, testCase.AntiAffinity)
		if err != nil {
			t.Error(err)
		}

		expectedChunkIDs := make([]datastore.ChunkID, 0)
		for idx, _ := range testCase.Distribution {
			expectedChunkIDs = append(expectedChunkIDs, file.Chunks.Chunks[idx].ID)
		}

		expectedNodeIDs := make([][]string, 0)
		for _, nodeIDsx := range testCase.Distribution {
			nodeIDs := make([]string, 0)
			for _, idx := range nodeIDsx {
				nodeIDs = append(nodeIDs, cloud.Network().Nodes[idx].ID)
			}
			expectedNodeIDs = append(expectedNodeIDs, nodeIDs)
		}

		t.Logf("case(%d) ChunkNodes %v.", i, cloud.Network().ChunkNodes)
		if len(testCase.Distribution) != len(cloud.Network().ChunkNodes) {
			t.Errorf("case(%d).Distribution got length %d; want %d", i, len(testCase.Distribution), len(cloud.Network().ChunkNodes))
		}
		for chunkIDx, _ := range testCase.Distribution {
			eChunkID := expectedChunkIDs[chunkIDx]
			eNodeIDs := expectedNodeIDs[chunkIDx]
			nodeIDs, ok := cloud.Network().ChunkNodes[eChunkID]
			if !ok {
				t.Errorf("case(%d).Distribution missing chunk %v", i, eChunkID)
			}
			if !reflect.DeepEqual(eNodeIDs, nodeIDs) {
				t.Errorf("case(%d).Distribution.chunk(%v) got nodes %v; want %v", i, eChunkID, nodeIDs, eNodeIDs)
			}
		}
	}
}

func TestDistributionError(t *testing.T) {

}
