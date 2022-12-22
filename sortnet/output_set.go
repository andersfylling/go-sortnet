package sortnet

import "math/bits"

type BinarySequence uint16

func (b BinarySequence) OnesCount() int {
	return bits.OnesCount16(uint16(b))
}

type OutputSet interface {
	Add(sequence BinarySequence)
}

func PopulateOutputSet(set OutputSet, sequenceSize int) OutputSet {
	limit := BinarySequence(0b111_111_111_111 >> (limit_sequence_size - sequenceSize))
	for seq := BinarySequence(0); seq < limit; seq++ {
		set.Add(seq)
	}

	return set
}
