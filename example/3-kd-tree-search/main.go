package main

import (
	"context"
	"fmt"
	"github.com/andersfylling/go-sortnet/sortnet"
	"github.com/andersfylling/go-sortnet/sortnet/outputset"
	"os"
	"os/signal"
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	defer func() {
		signal.Stop(c)
		cancel()
	}()
	go func() {
		select {
		case <-c:
			cancel()
		case <-ctx.Done():
		}
	}()

	workChan := make(chan *Work, 1000)
	lookupChannel := make(chan *LookupRequest, 1000)

	for i := 0; i < 10; i++ {
		go SearchWorker(ctx, workChan, lookupChannel)
	}

	rounds := 0
	for {
		fmt.Printf("Round %d\n", rounds)

		networks = GenerateNetworks(allComparators, networks)
		fmt.Printf("\tgenerated %d networks\n", len(networks))
		outputSets := GenerateOutputSets(networks)
		fmt.Println("\tgenerated output sets")

		tree := NewKDTree()
		points := CreateMetadataPoints(outputSets)
		tree.Insert(points)
		tree.Balance()
		fmt.Printf("\tbuilt kd tree with %d dimensions\n", len(points[0].Coordinates()))

		go Producer(tree, outputSets, workChan)
		WaitForResults(ctx, outputSets, lookupChannel)
		if ctx.Err() != nil {
			return
		}

		before := len(networks)
		// PruneSubsumedOutputSets(tree, outputSets)
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

func CreateMetadataPoints(sets []*outputset.PartitionedOrdered) []*Metadata {
	points := make([]*Metadata, 0, len(sets))
	for i, s := range sets {
		point := CreateMetadataPoint(s)
		point.Index = i

		points = append(points, point)
	}

	return points
}

func CreateMetadataPoint(set *outputset.PartitionedOrdered) *Metadata {
	var ones []int
	var zeros []int
	var sizes []int
	for pi := range set.Partitions {
		sizes = append(sizes, set.PartitionSize(pi))
		ones = append(ones, set.OnesMasks[pi].OnesCount())
		zeros = append(zeros, set.ZerosMasks[pi].OnesCount())
	}

	return &Metadata{
		Size:           set.Size(),
		Ones:           ones,
		Zeros:          zeros,
		PartitionSizes: sizes,
	}
}

func subsumes(a, b *outputset.PartitionedOrdered) bool {
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

func Producer(tree *KDTree, sets []*outputset.PartitionedOrdered, workChan chan *Work) {
	for currentID, currentSet := range sets {
		if currentSet == nil {
			continue
		}

		base := CreateMetadataPoint(currentSet)
		indexes := tree.FindCandidates(base)

		var workSets []SetReference
		for _, i := range indexes {
			b := sets[i]
			if b == currentSet || b == nil {
				continue
			}

			workSets = append(workSets, SetReference{
				set: b,
				id:  i,
			})
		}

		workChan <- &Work{
			current: SetReference{
				set: currentSet,
				id:  currentID,
			},
			targets: workSets,
		}
	}
}

type SetReference struct {
	set *outputset.PartitionedOrdered
	id  int
}

type Work struct {
	current SetReference
	targets []SetReference
}

type LookupRequest struct {
	channel chan *LookupResult
	ids     []int
	id      int
}

type LookupResult struct {
	set *outputset.PartitionedOrdered
	id  int
}

func WaitForResults(ctx context.Context, sets []*outputset.PartitionedOrdered, work chan *LookupRequest) {
	pendingWrites := map[int]*LookupRequest{}

	for {
		select {
		case request := <-work:
			pendingWrites[request.id] = request
		case <-ctx.Done():
			return
		}

		if len(pendingWrites) != len(sets) {
			continue
		}

		// delete marked sets in order
		for i := range sets {
			if sets[i] == nil {
				continue
			}

			w := pendingWrites[i]
			for _, id := range w.ids {
				sets[id] = nil
			}
		}
		break
	}

}

func SearchWorker(ctx context.Context, workChan chan *Work, lookupChannel chan<- *LookupRequest) {
	for {
		select {
		case <-ctx.Done():
			return
		case work, ok := <-workChan:
			if !ok {
				return
			}

			var pruned []int
			for _, target := range work.targets {
				if subsumes(work.current.set, target.set) {
					pruned = append(pruned, target.id)
				}
			}

			lookupChannel <- &LookupRequest{nil, pruned, work.current.id}
		}
	}
}
