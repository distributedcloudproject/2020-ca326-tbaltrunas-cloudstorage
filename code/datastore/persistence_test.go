package datastore

import (
	"cloud/network"
	"testing"
	"reflect"
	"hash"
	"hash/fnv"
)

func SampleDataStoreFile() File {
	// set up test file and its chunk
	h := hash.Hash(fnv.New32())
	h.Write([]byte("test data"))
	chunkID := FileChunkIDType(string(h.Sum(make([]byte, 0))))
	fileChunks := FileChunks{
		ChunkNumber: 1,
		ChunkSize: 9,
		ChunkIDs: []FileChunkIDType{chunkID},
	}

	file := File{
		Path: "/home/test",
		Size: 100,
		Chunks: fileChunks,
	}
	return file
}

func SampleDataStore(sampleFile File) DataStore {
	return DataStore{Files: []File{sampleFile}}
}

func SampleFileChunkLocations(sampleChunkID FileChunkIDType) FileChunkLocations {
	chunkLocations := make(FileChunkLocations)
	testNode := ChunkNodeType(network.Node{
		IP: "127.0.0.1",
		Name: "testnode",
	})
	chunkLocations[sampleChunkID] = []ChunkNodeType{testNode}
	return chunkLocations
}

func TestPersistDataStore(t *testing.T) {
	var err error

	// set up sample data structures
	file := SampleDataStoreFile()
	t.Logf("sample file: %v", file)
	chunkID := file.Chunks.ChunkIDs[0]
	t.Logf("chunkID hash: %v", chunkID)
	dataStore := SampleDataStore(file)
	t.Logf("data store: %v", dataStore)
	chunkLocations := SampleFileChunkLocations(chunkID)
	t.Logf("chunk locations: %v", chunkLocations)

	persistency_path := "/tmp/cloud_test_persistence"
	t.Logf("Saving at %s", persistency_path)
	err = Save(persistency_path, dataStore)
	if err != nil {
		t.Error(err)
	}
	var dataStore2 DataStore
	err = Load(persistency_path, &dataStore2)
	if err != nil {
		t.Error(err)
	}

	// Note that DeepEqual has arguments against using it.
	// https://stackoverflow.com/a/45222521
	// An alternative struct comparison method may be needed in the future.
	t.Logf("Original: %v", dataStore)
	t.Logf("Loaded: %v", dataStore2)
	if !reflect.DeepEqual(dataStore, dataStore2) {
		t.Errorf("Mismatch between original data store and loaded data store")
	}

	err = Save(persistency_path, chunkLocations)
	if err != nil {
		t.Error(err)
	}
	var chunkLocations2 FileChunkLocations
	err = Load(persistency_path, &chunkLocations2)
	if err != nil {
		t.Error(err)
	}
	t.Logf("Original: %v", chunkLocations)
	t.Logf("Loaded: %v", chunkLocations2)
	if !reflect.DeepEqual(chunkLocations, chunkLocations2) {
		t.Errorf("Mismatch between original chunk structure and loaded chunk structure.")
	}
}
