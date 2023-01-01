package main

import (
	"fmt"
	"github.com/andersfylling/go-sortnet/example"
	"github.com/andersfylling/go-sortnet/sortnet"
	"github.com/andersfylling/go-sortnet/sortnet/outputset"
)

// see example/configuration.go
const (
	Channels = example.Channels
	Workers  = example.Workers
)

// configurable variables
// see example/configuration.go
var (
	NewSet outputset.NewSet = example.NewSet
)

func init() {
	if example.PruningStrategy == example.ParallelPruning {
		panic("parallel pruning has not been implemented")
	}
}

func main() {
	allComparators := sortnet.AllComparatorCombinations(Channels)
	networks := []sortnet.Network{
		&sortnet.ComparatorNetwork{},
	}

	for round := 0; ; round++ {
		fmt.Printf("Round %d\n", round)

		networks = GenerateNetworks(allComparators, networks)
		fmt.Printf("\tgenerated %d networks\n", len(networks))
		outputSets := GenerateOutputSets(networks)
		fmt.Println("\tgenerated output sets")

		before := len(networks)
		PruneSubsumedOutputSets(outputSets)
		networks = NetworksWithNonNilOutputset(outputSets, networks)

		fmt.Printf("\tpruned %d networks - %d remaining\n", before-len(networks), len(networks))

		if len(networks) == 1 {
			break
		}
	}

	fmt.Printf("\n\n")
	fmt.Println("Discovered sorting network")
	fmt.Println(networks[0])
}

func NetworksWithNonNilOutputset(sets []sortnet.OutputSet, networks []sortnet.Network) []sortnet.Network {
	for i := range sets {
		if sets[i] == nil {
			networks[i] = nil
		}
	}

	for i := 0; i < len(networks); i++ {
		if networks[i] == nil {
			networks[i] = networks[len(networks)-1]
			networks = networks[:len(networks)-1]
			i--
		}
	}

	return networks
}

func GenerateNetworks(comparators []sortnet.Comparator, networks []sortnet.Network) []sortnet.Network {
	derivatives := make([]sortnet.Network, 0, len(networks)*(len(comparators)-1))
	for _, network := range networks {
		derivatives = append(derivatives, network.Derive(comparators)...)
	}

	return derivatives
}

func GenerateOutputSets(networks []sortnet.Network) []sortnet.OutputSet {
	sets := make([]sortnet.OutputSet, len(networks))
	for i, network := range networks {
		complete := NewSet(Channels)
		sets[i] = complete.Derive(network)
	}

	return sets
}

func PruneSubsumedOutputSets(sets []sortnet.OutputSet) {
	for _, a := range sets {
		if a == nil {
			continue
		}

		for bi, b := range sets {
			if b == nil || b == a {
				continue
			}

			if a.IsSubset(b, nil) {
				sets[bi] = nil
			}
		}
	}
}
