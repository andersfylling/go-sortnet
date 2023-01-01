package sortnet

func MapLegalPlacementsBitMap(constraints []BinarySequence, dst, src BinarySequence) {
	for i := range constraints {
		if ((src >> i) & 0b1) == 0 {
			continue
		}
		constraints[i] &= dst
	}
}

func PermutationBitMapPositions(channels int, dst, src *SetMetadata) []BinarySequence {
	// assert len(src.OnesMasks) == len(dst.OnesMasks)
	// assert len(src.ZerosMasks) == len(dst.ZerosMasks)

	partitions := len(src.OnesMasks)
	constraints := make([]BinarySequence, partitions)

	// set every relevant bit to start off with no constraints
	for i := range constraints {
		constraints[i] = SequenceMask(channels)
	}

	// as we identify relations between the two sets we start introducing constraints for
	// possible permutation maps
	for pi := 0; pi < partitions; pi++ {
		MapLegalPlacementsBitMap(constraints, dst.OnesMasks[pi], src.OnesMasks[pi])
		MapLegalPlacementsBitMap(constraints, dst.ZerosMasks[pi], src.ZerosMasks[pi])

		if dst.PartitionSizes[pi] == src.PartitionSizes[pi] {
			MapLegalPlacementsBitMap(constraints, ^dst.OnesMasks[pi], ^src.OnesMasks[pi])
			MapLegalPlacementsBitMap(constraints, ^dst.ZerosMasks[pi], ^src.ZerosMasks[pi])
		}
	}

	return constraints
}

// QuickValidatePermutationBitMapPositions verifies:
// 1. that any position can be moved to at least one position
// 2. that the combined legal positions will at least once touch every position
func QuickValidatePermutationBitMapPositions(channels int, constraints []BinarySequence) bool {
	var output BinarySequence
	for _, options := range constraints {
		if options == 0 {
			return false
		}

		output |= options
	}

	return output == SequenceMask(channels)
}

// GeneratePermutationsByBitmap will identify all plausible permutations using backtracking.
func GeneratePermutationsByBitmap(channels int, src, dst *SetMetadata, hook PermutationGeneratorHook) bool {
	reject := func(permutationMap PermutationMap) bool {
		// check if a position is referenced twice
		var mask BinarySequence
		for _, offset := range permutationMap {
			bit := BinarySequence(0b1 << offset)
			if bit&mask != 0 {
				return true
			}
			mask |= bit
		}

		return false
	}

	constraints := PermutationBitMapPositions(channels, dst, src)
	if !QuickValidatePermutationBitMapPositions(channels, constraints) {
		return false
	}

	// each element represents a permuted bit position. You need N elements for a binary sequence of size N to create
	// a complete permutation map.
	var stack []PermutationMap

	// create the first legal positions based on the positional constraints
	for it := NewSequenceIterator(constraints[0]); !it.Empty(); {
		offset := it.Next()
		stack = append(stack, PermutationMap{offset})
	}

	for len(stack) != 0 {
		candidate := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if reject(candidate) {
			continue
		}

		// extract finished permutations
		if len(candidate) == channels {
			if hook(candidate) {
				return true
			}
			continue
		}

		position := len(candidate)
		for it := NewSequenceIterator(constraints[position]); !it.Empty(); {
			offset := it.Next()

			newCandidate := make(PermutationMap, len(candidate), len(candidate)+1)
			copy(newCandidate, candidate)
			newCandidate = append(newCandidate, offset)

			stack = append(stack, newCandidate)
		}
	}

	return false
}
