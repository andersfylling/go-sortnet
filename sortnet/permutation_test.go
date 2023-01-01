package sortnet

import "testing"

func TestFunctionSignatures(t *testing.T) {
	_ = []GeneratePermutationsFunc{
		GeneratePermutationsByBitmap,
	}
}
