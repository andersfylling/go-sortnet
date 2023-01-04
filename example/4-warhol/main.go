package main

import (
	"fmt"
	"github.com/andersfylling/go-sortnet/example"
	"github.com/andersfylling/go-sortnet/sortnet"
	"github.com/andersfylling/go-sortnet/sortnet/outputset"
	"github.com/cheggaaa/pb/v3"
	"sync/atomic"
	"time"
)

// see example/configuration.go
const (
	Channels     = example.Channels
	Workers      = example.Workers
	FiltersLimit = example.FiltersLimit
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

var programStart = time.Now()

func run() {
	defer func() {
		fmt.Printf("program finished in %s\n", time.Now().Sub(programStart))
	}()
	allComparators := sortnet.AllComparatorCombinations(Channels)
	networks := []sortnet.Network{
		&sortnet.ComparatorNetwork{},
	}

	k := 0
	for {
		fmt.Println()
		fmt.Printf("Round %d\n", k+1)
		k++

		var sets []sortnet.OutputSet
		networks, sets = Generate(allComparators, networks)

		sets = Prune(sets)

		before := len(networks)
		networks = NetworksWithNonNilOutputset(sets, networks)
		fmt.Printf("\tpruned %d networks - %d remaining\n", before-len(networks), len(networks))

		if len(networks) == 1 && k > 1 || k == FiltersLimit {
			break
		}
	}

	if len(networks) == 1 {
		fmt.Println("Sorting network discovered")
		fmt.Println(networks[0])
	}
}

func NetworksWithNonNilOutputset(sets []sortnet.OutputSet, networks []sortnet.Network) []sortnet.Network {
	selected := make([]sortnet.Network, 0, len(networks))
	for i := range sets {
		if sets[i] != nil {
			id := sets[i].Metadata().NetworkID
			selected = append(selected, networks[int(id)])
		}
	}

	return selected
}

type GenerateWork struct {
	network sortnet.Network
	set     sortnet.OutputSet
}

func Generate(comparators []sortnet.Comparator, networks []sortnet.Network) ([]sortnet.Network, []sortnet.OutputSet) {
	bar := pb.StartNew(len(networks))
	defer bar.Finish()
	defer bar.SetCurrent(bar.Total())

	bar.SetMaxWidth(100)
	bar.SetRefreshRate(100 * time.Millisecond)

	workChan := make(chan *GenerateWork, Workers)

	go func() {
		defer close(workChan)
		for _, network := range networks {
			workChan <- &GenerateWork{network: network}
		}
	}()

	mergeChan := make(chan *GenerateWork, Workers)
	active := int32(Workers)
	for i := 0; i < Workers; i++ {
		go func() {
			tree := &Container{}
			var children []*GenerateWork
			var localID sortnet.NetworkID

			for work := range workChan {
				set := NewSet(Channels)
				set = set.Derive(work.network)

				left := localID
				for _, child := range work.network.Derive(comparators) {
					childSet := NewSet(Channels)
					childSet = childSet.Derive(child)
					childSet.Metadata().NetworkID = localID

					if set.Size() == childSet.Size() && set.IsSubset(childSet, nil) {
						continue
					}

					children = append(children, &GenerateWork{
						network: child,
						set:     childSet,
					})
					tree.Insert([]sortnet.OutputSet{childSet})
					localID++
				}

				tree.Balance()
				for id := int(left); id < len(tree.sets); id++ {
					if tree.sets[id] == nil {
						continue
					}
					tree.Prune(PruneParallel(id, tree.sets, tree))
				}

				if tree.dirty > 1000 {
					tree.Rebalance()
				}
				bar.Increment()
			}

			for id := range tree.sets {
				if tree.sets[id] == nil {
					continue
				}
				mergeChan <- children[id]
			}

			if atomic.AddInt32(&active, -1) == 0 {
				close(mergeChan)
			}
		}()
	}

	var derivatives []sortnet.Network
	var sets []sortnet.OutputSet
	var id sortnet.NetworkID
	for work := range mergeChan {
		work.set.Metadata().NetworkID = id
		id++

		derivatives = append(derivatives, work.network)
		sets = append(sets, work.set)
	}

	return derivatives, sets
}

func CreateMetadataPoint(setAbstraction sortnet.OutputSet) *Metadata {
	md := setAbstraction.Metadata()

	ones := make([]int, 0, len(md.OnesMasks))
	zeros := make([]int, 0, len(md.ZerosMasks))
	sizes := make([]int, 0, len(md.PartitionSizes))
	for pi := 0; pi < len(md.PartitionSizes); pi++ {
		sizes = append(sizes, md.PartitionSizes[pi])
		ones = append(ones, md.OnesMasks[pi].OnesCount())
		zeros = append(zeros, md.ZerosMasks[pi].OnesCount())
	}

	return &Metadata{
		Size:           md.Size,
		Ones:           ones,
		Zeros:          zeros,
		PartitionSizes: sizes,
		ID:             md.NetworkID,
	}
}

func subsumesByPermutation(a, b sortnet.OutputSet) bool {
	// ST1, ST2, ST3 have moved into the KD-tree, see CreateMetadataPoint
	return GeneratePermutations(Channels, a.Metadata(), b.Metadata(), func(permutationMap sortnet.PermutationMap) bool {
		return a.IsSubset(b, permutationMap)
	})
}

func Prune(sets []sortnet.OutputSet) []sortnet.OutputSet {
	const rebuildAfter = 2000

	bar := pb.StartNew(len(sets))
	defer bar.Finish()
	defer bar.SetCurrent(bar.Total())

	bar.SetMaxWidth(100)
	bar.SetRefreshRate(100 * time.Millisecond)

	tree := &Container{}
	tree.Insert(sets)
	tree.Balance()

	for currentID := range tree.sets {
		if tree.sets[currentID] == nil {
			continue
		}
		bar.SetCurrent(int64(currentID))

		subsumed := PruningStrategy(currentID, sets, tree)
		tree.Prune(subsumed)

		if currentID%rebuildAfter == 0 {
			tree.Rebalance()
		}
	}

	return tree.sets
}

type PruneMethod = func(currentID int, sets []sortnet.OutputSet, tree *Container) []sortnet.NetworkID

func PruneSerial(currentID int, sets []sortnet.OutputSet, tree *Container) []sortnet.NetworkID {
	panic("wip")
	//set := sets[currentID]
	//targets := tree.Search(set)
	//
	//var ids []sortnet.NetworkID
	//for _, target := range targets {
	//	if subsumesByPermutation(set, target) {
	//		ids = append(ids, target.Metadata().NetworkID)
	//	}
	//}
	//
	//return ids
}

type Work struct {
	set sortnet.OutputSet
	id  sortnet.NetworkID
}

func PruneParallel(currentID int, sets []sortnet.OutputSet, tree *Container) []sortnet.NetworkID {
	workChan := make(chan *Work, Workers)

	go func() {
		defer close(workChan)
		tree.SearchParallel(sets[currentID], workChan)
	}()

	subsumedChan := make(chan sortnet.NetworkID)
	active := int32(Workers)
	for i := 0; i < Workers; i++ {
		go func() {
			for work := range workChan {
				if subsumesByPermutation(sets[currentID], work.set) {
					subsumedChan <- work.id
				}
			}

			if atomic.AddInt32(&active, -1) == 0 {
				close(subsumedChan)
			}
		}()
	}

	var ids []sortnet.NetworkID
	for id := range subsumedChan {
		ids = append(ids, id)
	}
	return ids
}
