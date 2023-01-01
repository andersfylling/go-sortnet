package example

import (
	"fmt"
	"github.com/andersfylling/go-sortnet/sortnet"
	"github.com/andersfylling/go-sortnet/sortnet/outputset"
)

// configurable constants
const (
	// Channels also known as "N", sets the number of network channels or the sequence length.
	Channels = 7
	Workers  = 16
)

// configurable variables
var (
	GeneratePermutations sortnet.GeneratePermutationsFunc = sortnet.GeneratePermutationsByBitmap
	NewSet               outputset.NewSet                 = outputset.NewPartitionedOrdered
	PruningStrategy      PruningStrategyType              = ParallelPruning
)

type PruningStrategyType int

const (
	ParallelPruning PruningStrategyType = iota
	SerialPruning
)

func init() {
	fmt.Println()
	fmt.Println("###############################################")
	fmt.Println("###############################################")
	fmt.Println("####                                       ####")
	fmt.Println("####   Proof based algorithm for finding   ####")
	fmt.Println("####  minimum number of comparators for a  ####")
	fmt.Printf("####    sorting network with %d channels.   ####\n", Channels)
	fmt.Println("####                                       ####")
	fmt.Printf("#### Parallelism: %t                     ####\n", PruningStrategy == ParallelPruning)
	fmt.Printf("#### Workers: %d                           ####\n", Workers)
	fmt.Println("####                                       ####")
	fmt.Println("###############################################")
	fmt.Println("###############################################")
	fmt.Println()
	fmt.Println()
}
