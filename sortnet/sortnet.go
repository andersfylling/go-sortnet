package sortnet

import "math/bits"

const LimitSequenceSize = 16
const LimitSequenceMask = 0b1111_1111_1111_1111
const None = 0

type BinarySequence uint16

func (b BinarySequence) OnesCount() int {
	return bits.OnesCount16(uint16(b))
}

// MostSignificantBitOffset returns offset for most significant set bit, -1 when there are no set bits.
func (b BinarySequence) MostSignificantBitOffset() int {
	width := bits.Len16(uint16(b))
	return width - 1
}

func SequenceMask(sequenceSize int) BinarySequence {
	return BinarySequence(LimitSequenceMask >> (LimitSequenceSize - sequenceSize))
}

func NewSequenceIterator(seq BinarySequence) *SequenceIterator {
	return &SequenceIterator{
		seq: seq,
	}
}

type SequenceIterator struct {
	seq BinarySequence
}

func (it *SequenceIterator) Empty() bool {
	return it.seq == 0
}

// Next returns the next most significant bit (offset). Returns -1 when there are no set bits / done iterating.
// You must call Empty() to avoid underflow scenarios, as this will run forever otherwise.
func (it *SequenceIterator) Next() int {
	offset := it.seq.MostSignificantBitOffset()
	it.seq ^= 0b1 << BinarySequence(offset)
	return offset
}
