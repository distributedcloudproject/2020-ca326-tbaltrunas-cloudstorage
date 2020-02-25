package network

// NodeBenchmark represents a set of benchmarks for a node given by ID.
type NodeBenchmark struct {
	ID string
	StorageSpaceRemaining uint64
}

func (n *cloudNode) Benchmark() (NodeBenchmark, error) {
	benchmarks := NodeBenchmark{ID: n.ID}
	storageSpaceRemaining, err := n.StorageSpaceRemaining()
	if err != nil {
		return benchmarks, err
	}
	benchmarks.StorageSpaceRemaining = storageSpaceRemaining
	return benchmarks, nil
}
