package sortnet

// PermutationMap delegates positional transformation for any binary sequence.
//
//	given the map (m) [0, 1, 2, 3], no transformation is required, as each index is in order
//	meaning applying this permutation would not actually change the binary sequence.
//
//	Given the map [0, 2, 1, 3] the 2th and 1th position would swap places, while 0th and 3th would
//	remain unchanged.
//
// Given the binary position X of a sequence, it's bit value should be swapped with bit position m[x] of the permutation map.
type PermutationMap []int

func ApplyPermutation(src BinarySequence, permutation PermutationMap) BinarySequence {
	var dst BinarySequence
	for i := 0; i < len(permutation); i++ {
		value := (src >> i) & 0b1
		dst |= value << permutation[i]
	}

	return dst
}
