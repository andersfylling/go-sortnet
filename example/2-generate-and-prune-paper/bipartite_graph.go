package main

import (
	"github.com/andersfylling/go-sortnet/sortnet"
	"github.com/andersfylling/go-sortnet/sortnet/outputset"
)

func Backtrack(constraints []sortnet.BinarySequence, cb func(permutationMap sortnet.PermutationMap) bool) bool {
	reject := func(permutationMap sortnet.PermutationMap) bool {
		// check if a position is referenced twice
		var mask sortnet.BinarySequence
		for _, offset := range permutationMap {
			bit := sortnet.BinarySequence(0b1 << offset)
			if bit&mask == 1 {
				return true
			}
			mask |= bit
		}

		return false
	}

	// each element represents a permuted bit position. You need N elements for a binary sequence of size N to create
	// a complete permutation map.
	var stack []sortnet.PermutationMap

	// create the first legal positions based on the positional constraints
	for it := sortnet.NewSequenceIterator(constraints[0]); !it.Empty(); {
		offset := it.Next()
		stack = append(stack, sortnet.PermutationMap{offset})
	}

	for len(stack) != 0 {
		candidate := stack[len(stack)-1] // fifo
		stack = stack[:len(stack)-1]

		if reject(candidate) {
			continue
		}

		// extract finished permutations
		if len(candidate) == N {
			if cb(candidate) {
				return true
			}
			continue
		}

		position := len(candidate)
		for it := sortnet.NewSequenceIterator(constraints[position]); !it.Empty(); {
			offset := it.Next()

			newCandidate := make(sortnet.PermutationMap, len(candidate), len(candidate)+1)
			copy(newCandidate, candidate)
			newCandidate = append(newCandidate, offset)

			stack = append(stack, newCandidate)
		}
	}

	return false
}

func MapRelation(constraints []sortnet.BinarySequence, dst, src sortnet.BinarySequence) {
	for i := range constraints {
		if ((src >> i) & 0b1) == 0 {
			continue
		}
		constraints[i] &= dst
	}
}

func Constraints(dst, src *outputset.Metadata) []sortnet.BinarySequence {
	// assert len(src.OnesMasks) == len(dst.OnesMasks)
	// assert len(src.ZerosMasks) == len(dst.ZerosMasks)

	partitions := len(src.OnesMasks)
	constraints := make([]sortnet.BinarySequence, partitions)

	// set every relevant bit to start off with no constraints
	for i := range constraints {
		constraints[i] = sortnet.SequenceMask(N)
	}

	// as we identify relations between the two sets we start introducing constraints for
	// possible permutation maps
	for pi := 0; pi < partitions; pi++ {
		MapRelation(constraints, dst.OnesMasks[pi], src.OnesMasks[pi])
		MapRelation(constraints, dst.ZerosMasks[pi], src.ZerosMasks[pi])
	}

	return constraints
}

func ValidateConstraintsFast(constraints []sortnet.BinarySequence) bool {
	var output sortnet.BinarySequence
	for _, options := range constraints {
		if options == 0 {
			return false
		}

		output |= options
	}

	return output == sortnet.SequenceMask(len(constraints))
}
