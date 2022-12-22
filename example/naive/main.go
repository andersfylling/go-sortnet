package main

import (
	"fmt"
	"github.com/andersfylling/go-sortnet/sortnet"
)

const N = 5

func main() {
	allComparators := sortnet.AllComparatorCombinations(N)
	networks := []sortnet.Network{
		&sortnet.ComparatorNetwork{},
	}

	rounds := 0
	for {
		fmt.Printf("Round %d\n", rounds)

		networks = GenerateNetworks(allComparators, networks)
		fmt.Printf("\tgenerated %d networks\n", len(networks))
		outputSets := GenerateOutputSets(networks)
		fmt.Println("\tgenerated output sets")

		before := len(networks)
		PruneSubsumedOutputSets(outputSets)

		for i := range outputSets {
			if outputSets[i] == nil {
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
		fmt.Printf("\tpruned %d networks - %d remaining\n", before-len(networks), len(networks))

		if len(networks) == 1 {
			break
		}

		if rounds > 50 {
			break
		}
		rounds++
	}

	var sortingNetwork sortnet.Network
	for _, network := range networks {
		if network != nil {
			sortingNetwork = network
		}
	}

	fmt.Println("Network")
	fmt.Println(sortingNetwork)
}

func GenerateNetworks(comparators []sortnet.Comparator, networks []sortnet.Network) []sortnet.Network {
	derivatives := make([]sortnet.Network, 0, len(networks))

	for _, network := range networks {
		if network == nil {
			continue
		}

		derivatives = append(derivatives, network.Derive(comparators)...)
	}

	return derivatives
}

func GenerateOutputSets(networks []sortnet.Network) []*SliceOutputSet {
	derivatives := make([]*SliceOutputSet, len(networks))

	for i, network := range networks {
		complete := NewSliceOutputSet(N)
		derivatives[i] = complete.Derive(network)
	}

	return derivatives
}

func PruneSubsumedOutputSets(sets []*SliceOutputSet) {
	subsumes := func(a, b *SliceOutputSet) bool {
		return a.IsSubset(b)
	}

	for i := range sets {
		a := sets[i]
		if a == nil {
			continue
		}

		for j := i + 1; j < len(sets); j++ {
			b := sets[j]
			if b == nil {
				continue
			}

			if subsumes(a, b) {
				sets[j] = nil
			} else if subsumes(b, a) {
				sets[i] = nil
				break
			}
		}
	}
}
