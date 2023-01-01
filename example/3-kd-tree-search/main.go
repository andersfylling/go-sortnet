package main

import (
	"context"
	"fmt"
	"github.com/andersfylling/go-sortnet/example"
	"github.com/andersfylling/go-sortnet/sortnet"
	"github.com/andersfylling/go-sortnet/sortnet/outputset"
	"github.com/cheggaaa/pb/v3"
	"golang.org/x/sync/errgroup"
	"sync/atomic"
	"time"
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

	k := 0
	for {
		fmt.Printf("Round %d\n", k)
		k++

		var sets []sortnet.OutputSet
		networks, sets = Generate(allComparators, networks)
		fmt.Printf("\tgenerated %d networks & output sets\n", len(networks))

		sets = Prune(sets)

		before := len(networks)
		networks = NetworksWithNonNilOutputset(sets, networks)
		fmt.Printf("\tpruned %d networks - %d remaining\n", before-len(networks), len(networks))

		if len(networks) == 1 && k > 1 || k > 50 {
			break
		}
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

func Generate(comparators []sortnet.Comparator, networks []sortnet.Network) ([]sortnet.Network, []sortnet.OutputSet) {
	derivatives := make([]sortnet.Network, 0, len(networks))
	sets := make([]sortnet.OutputSet, 0, len(networks))

	for _, network := range networks {
		set := NewSet(Channels)
		set = set.Derive(network)

		for _, child := range network.Derive(comparators) {
			childSet := NewSet(Channels)
			childSet = childSet.Derive(child)

			if subsumes(set, childSet) {
				continue
			}

			derivatives = append(derivatives, child)
			sets = append(sets, childSet)
		}
	}

	return derivatives, sets
}

func CreateMetadataPoints(sets []sortnet.OutputSet) []*Metadata {
	points := make([]*Metadata, 0, len(sets))
	for i, s := range sets {
		if s == nil {
			continue
		}
		point := CreateMetadataPoint(s)
		point.Index = i

		points = append(points, point)
	}

	return points
}

func CreateMetadataPoint(setAbstraction sortnet.OutputSet) *Metadata {
	md := setAbstraction.Metadata()

	var ones []int
	var zeros []int
	var sizes []int
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
	}
}

func subsumes(a, b sortnet.OutputSet) bool {
	if a.Size() > b.Size() {
		return false
	}

	return a.IsSubset(b, nil)
}

func subsumesByPermutation(a, b sortnet.OutputSet) bool {
	// ST1, ST2, ST3 have moved into the KD-tree, see CreateMetadataPoint
	return GeneratePermutations(Channels, a.Metadata(), b.Metadata(), func(permutationMap sortnet.PermutationMap) bool {
		return a.IsSubset(b, permutationMap)
	})
}

func Prune(sets []sortnet.OutputSet) []sortnet.OutputSet {
	const rebuildAfter = 1000

	var tree *KDTree
	rebuild := func() {
		tree = NewKDTree()
		points := CreateMetadataPoints(sets)
		tree.Insert(points)
		tree.Balance()
	}

	bar := pb.StartNew(len(sets))
	defer bar.Finish()

	bar.SetMaxWidth(150)
	bar.SetRefreshRate(100 * time.Millisecond)

	var pruned int
	rebuild()
	for currentID, set := range sets {
		bar.Increment()
		if set == nil {
			continue
		}

		subsumed := PruningStrategy(currentID, sets, tree)
		shouldRebuild := false
		for _, id := range subsumed {
			sets[id] = nil
			pruned++

			if pruned%rebuildAfter == 0 {
				shouldRebuild = true
			}
		}

		if shouldRebuild {
			rebuild()
		}
	}

	return sets
}

type PruneMethod = func(currentID int, sets []sortnet.OutputSet, tree *KDTree) []int

func PruneSerial(currentID int, sets []sortnet.OutputSet, tree *KDTree) []int {
	base := CreateMetadataPoint(sets[currentID])
	indexes := tree.FindCandidates(base)

	var ids []int
	for _, i := range indexes {
		target := sets[i]
		if currentID == i || target == nil {
			continue
		}

		if subsumesByPermutation(sets[currentID], target) {
			ids = append(ids, i)
		}
	}

	return ids
}

type Work struct {
	set sortnet.OutputSet
	id  int
}

func PruneParallel(currentID int, sets []sortnet.OutputSet, tree *KDTree) []int {
	g, _ := errgroup.WithContext(context.Background())
	workChan := make(chan *Work, Workers)

	g.Go(func() error {
		defer close(workChan)
		base := CreateMetadataPoint(sets[currentID])
		indexes := tree.FindCandidates(base)

		for _, i := range indexes {
			target := sets[i]
			if currentID == i || target == nil {
				continue
			}

			workChan <- &Work{
				set: target,
				id:  i,
			}
		}
		return nil
	})

	subsumedChan := make(chan int)
	active := int32(Workers)
	for i := 0; i < Workers; i++ {
		g.Go(func() error {
			for work := range workChan {
				if subsumesByPermutation(sets[currentID], work.set) {
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
