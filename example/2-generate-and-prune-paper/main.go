package main

import (
	"fmt"
	"github.com/andersfylling/go-sortnet/sortnet"
	"github.com/andersfylling/go-sortnet/sortnet/outputset"
)

const N = 8

func main() {
	run()
}

func run() {
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
		networks = NetworksWithNonNilOutputset(outputSets, networks)
		fmt.Printf("\tpruned %d networks - %d remaining\n", before-len(networks), len(networks))

		if len(networks) == 1 && rounds > 1 {
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

func NetworksWithNonNilOutputset(sets []*outputset.PartitionedOrdered, networks []sortnet.Network) []sortnet.Network {
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
	derivatives := make([]sortnet.Network, 0, len(networks))

	for _, network := range networks {
		if network == nil {
			continue
		}

		derivatives = append(derivatives, network.Derive(comparators)...)
	}

	return derivatives
}

func GenerateOutputSets(networks []sortnet.Network) []*outputset.PartitionedOrdered {
	derivatives := make([]*outputset.PartitionedOrdered, len(networks))

	for i, network := range networks {
		complete := outputset.NewPartitionedOrdered(N)
		derivatives[i] = complete.Derive(network)
	}

	return derivatives
}

func PruneSubsumedOutputSets(sets []*outputset.PartitionedOrdered) {
	subsumes := func(a, b *outputset.PartitionedOrdered) bool {
		if a.Size() > b.Size() {
			return false
		}

		// permutation preconditions
		for p := range a.Partitions {
			if len(a.Partitions[p]) > len(b.Partitions[p]) {
				return false
			}

			if a.OnesMasks[p].OnesCount() > b.OnesMasks[p].OnesCount() {
				return false
			}
			if a.ZerosMasks[p].OnesCount() > b.ZerosMasks[p].OnesCount() {
				return false
			}
		}

		permutationConstraints := Constraints(b.Metadata, a.Metadata)
		if !ValidateConstraintsFast(permutationConstraints) {
			return false
		}

		return Backtrack(permutationConstraints, func(permutationMap sortnet.PermutationMap) bool {
			return a.IsSubset(b, permutationMap)
		})
	}

	for _, a := range sets {
		if a == nil {
			continue
		}

		for bi, b := range sets {
			if b == nil || b == a {
				continue
			}

			if subsumes(a, b) {
				sets[bi] = nil
			}
		}
	}
}
