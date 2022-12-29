package main

import (
	"fmt"
	"github.com/andersfylling/go-sortnet/sortnet"
	"github.com/andersfylling/go-sortnet/sortnet/outputset"
)

const N = 8

func main() {
	allComparators := sortnet.AllComparatorCombinations(N)
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

func NetworksWithNonNilOutputset(sets []*outputset.Unordered, networks []sortnet.Network) []sortnet.Network {
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

func GenerateOutputSets(networks []sortnet.Network) []*outputset.Unordered {
	sets := make([]*outputset.Unordered, len(networks))
	for i, network := range networks {
		complete := outputset.NewUnordered(N)
		sets[i] = complete.Derive(network)
	}

	return sets
}

func PruneSubsumedOutputSets(sets []*outputset.Unordered) {
	for _, a := range sets {
		if a == nil {
			continue
		}

		for bi, b := range sets {
			if b == nil || b == a {
				continue
			}

			if a.IsSubset(b) {
				sets[bi] = nil
			}
		}
	}
}
