package datastore

import (
	"cloud/network"
	"testing"
	"hash"
	"hash/fnv"
	"reflect"
)

func TestPersistDataStore(t *testing.T) {
	var err error

	// set up test file and its chunk
	h := hash.Hash(fnv.New32())
	h.Write([]byte("test data"))
	chunkID := FileChunkIDType(string(h.Sum(make([]byte, 0))))
	t.Logf("chunkID hash: %v", chunkID)
	
	file := File{
		Path: "/home/test",
		Size: 100,
		ChunkIDs: []FileChunkIDType{chunkID},
	}

	// keep track of file in data structures
	dataStore := DataStore{Files: []File{file}}
	t.Logf("data store: %v", dataStore)
	chunkLocations := make(FileChunkLocations)
	testNode := ChunkNodeType(network.Node{
		IP: "127.0.0.1",
		Name: "testnode",
	})
	chunkLocations[chunkID] = []ChunkNodeType{testNode}
	t.Logf("chunk locations: %v", chunkLocations)

	persistency_path := "/tmp/cloud_test_persistence"
	err = Save(persistency_path, dataStore)
	if err != nil {
		t.Error(err)
	}
	var dataStore2 DataStore
	err = Load(persistency_path, dataStore2)
	if err != nil {
		t.Error(err)
	}

	// Note that DeepEqual has arguments against using it.
	// https://stackoverflow.com/a/45222521
	// An alternative struct comparison method may be needed in the future.
	if !reflect.DeepEqual(dataStore, dataStore2) {
		t.Errorf("Mismatch between original data store and loaded data store. Original: %v. Loaded: %v", dataStore, dataStore2)
	}

	err = Save(persistency_path, chunkLocations)
	if err != nil {
		t.Error(err)
	}
	var chunkLocations2 FileChunkLocations
	err = Load(persistency_path, chunkLocations2)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(chunkLocations, chunkLocations2) {
		t.Errorf("Mismatch between original chunk structure and loaded chunk structure. Original: %v. Loaded: %v", chunkLocations, chunkLocations2)
	}
}
