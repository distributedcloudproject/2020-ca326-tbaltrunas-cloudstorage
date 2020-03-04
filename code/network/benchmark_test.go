package network

import (
	"testing"
)

func TestBenchmarkNetworkLatency(t *testing.T) {
	clouds, err := CreateTestClouds(2)
	if err != nil {
		t.Fatal(err)
	}
	cloud := clouds[0]
	for i, n := range cloud.Network().Nodes {
		t.Logf("Node %d: %v.", i + 1, n.ID)
	}
	other := cloud.GetCloudNode(cloud.Network().Nodes[1].ID)

	latency, err := other.NetworkLatency()
	if err != nil {
		t.Error(err)
	}
	t.Logf("Latency: %v", latency)
	if !(0 < latency) {
		t.Errorf("Something is wrong with latency: %v (should be non-zero)", latency)
	}

	// TODO: ability to pass custom ping function to NetworkLatency
	// TODO: test with a ping function that has a certain sleep delay
}
