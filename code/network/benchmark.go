package network

import (
	"time"
)

// FIXME: use plural of benchmark for name
// NodeBenchmark represents a set of benchmarks for a node given by ID.
type NodeBenchmark struct {
	ID string
	StorageSpaceRemaining uint64
	Latency time.Duration 
}

// TODO: merge this with NodeBenchmark
// CloudBenchmarkState represents the running benchmarks for this cloud's node.
type CloudBenchmarkState struct {
	StorageSpaceUsed uint64 // in bytes, how much storage is used already.
}

// Benchmark retrieves the benchmarks of the given node.
func (n *cloudNode) Benchmark() (NodeBenchmark, error) {
	benchmarks := NodeBenchmark{ID: n.ID}
	storageSpaceRemaining, err := n.StorageSpaceRemaining()
	if err != nil {
		return benchmarks, err
	}
	benchmarks.StorageSpaceRemaining = storageSpaceRemaining

	latency, err := n.NetworkLatency()
	if err != nil {
		return benchmarks, err
	}
	benchmarks.Latency = latency
	return benchmarks, nil
}

func (c *cloud) BenchmarkState() CloudBenchmarkState {
	c.Mutex.RLock()
	defer c.Mutex.RUnlock()
	return c.benchmarkState
}
	
func (c *cloud) SetBenchmarkState(benchmarkState CloudBenchmarkState) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	c.benchmarkState = benchmarkState
}
