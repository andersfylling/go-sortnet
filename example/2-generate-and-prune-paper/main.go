package main

import (
	"context"
	"fmt"
	"github.com/andersfylling/go-sortnet/example"
	"github.com/andersfylling/go-sortnet/sortnet"
	"github.com/andersfylling/go-sortnet/sortnet/outputset"
	"golang.org/x/sync/errgroup"
	"sync/atomic"
)

// see example/configuration.go
const (
	Channels = example.Channels
	Workers  = example.Workers
)

// configurable variables
// see example/configuration.go
var (
	GeneratePermutations sortnet.GeneratePermutationsFunc = example.GeneratePermutations
	NewSet               outputset.NewSet                 = example.NewSet
	PruningStrategy      PruneMethod                      = PruneSerial
)

func init() {
	if example.PruningStrategy == example.ParallelPruning {
		PruningStrategy = PruneParallel
	}
}

func main() {
	run()
}

func run() {
	allComparators := sortnet.AllComparatorCombinations(Channels)
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

		Prune(outputSets)

		before := len(networks)
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
	derivatives := make([]sortnet.Network, 0, len(networks))

	for _, network := range networks {
		if network == nil {
			continue
		}

		derivatives = append(derivatives, network.Derive(comparators)...)
	}

	return derivatives
}

func GenerateOutputSets(networks []sortnet.Network) []sortnet.OutputSet {
	derivatives := make([]sortnet.OutputSet, len(networks))

	for i, network := range networks {
		complete := NewSet(Channels)
		derivatives[i] = complete.Derive(network)
	}

	return derivatives
}

func SubsumptionTest(a, b *sortnet.SetMetadata) bool {
	return a.ST1(b) && a.ST2(b) && a.ST3(b)
}

func Subsumes(a, b sortnet.OutputSet) bool {
	return GeneratePermutations(Channels, a.Metadata(), b.Metadata(), func(permutationMap sortnet.PermutationMap) bool {
		return a.IsSubset(b, permutationMap)
	})
}

type PruneMethod = func(currentID int, sets []sortnet.OutputSet) []int

func Prune(sets []sortnet.OutputSet) {
	for currentID, a := range sets {
		if a == nil {
			continue
		}

		subsumed := PruningStrategy(currentID, sets)
		for _, id := range subsumed {
			sets[id] = nil
		}
	}
}

func PruneSerial(currentID int, sets []sortnet.OutputSet) []int {
	var ids []int
	for id, target := range sets {
		if currentID == id || target == nil {
			continue
		}

		if !SubsumptionTest(sets[currentID].Metadata(), target.Metadata()) {
			continue
		}

		if Subsumes(sets[currentID], target) {
			ids = append(ids, id)
		}
	}

	return ids
}

type Work struct {
	set sortnet.OutputSet
	id  int
}

func PruneParallel(currentID int, sets []sortnet.OutputSet) []int {
	g, _ := errgroup.WithContext(context.Background())
	workChan := make(chan *Work, Workers)

	g.Go(func() error {
		defer close(workChan)

		for id, target := range sets {
			if currentID == id || target == nil {
				continue
			}

			if !SubsumptionTest(sets[currentID].Metadata(), target.Metadata()) {
				continue
			}

			workChan <- &Work{
				set: target,
				id:  id,
			}
		}
		return nil
	})

	subsumedChan := make(chan int)
	active := int32(Workers)
	for i := 0; i < Workers; i++ {
		g.Go(func() error {
			for work := range workChan {
				if Subsumes(sets[currentID], work.set) {
					subsumedChan <- work.id
				}
			}

			if atomic.AddInt32(&active, -1) == 0 {
				close(subsumedChan)
			}
			return nil
		})
	}

	subsumed := map[int]bool{}
	g.Go(func() error {
		for id := range subsumedChan {
			subsumed[id] = true
		}
		return nil
	})
	_ = g.Wait()

	var ids []int
	for id := range subsumed {
		ids = append(ids, id)
	}
	return ids
}
