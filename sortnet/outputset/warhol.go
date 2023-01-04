package outputset

import (
	"github.com/andersfylling/go-sortnet/sortnet"
	"github.com/kelindar/bitmap"
)

func NewEmptyWarhol(channels int) *Warhol {
	set := &Warhol{
		SetMetadata: &sortnet.SetMetadata{},
	}
	for i := 0; i < channels; i++ {
		set.Channels = append(set.Channels, bitmap.Bitmap{})
	}

	return set
}

func NewWarhol(channels int) sortnet.OutputSet {
	return sortnet.PopulateOutputSet(NewEmptyWarhol(channels), channels)
}

// Warhol is optimized for applying permutations. Instead of having a list of binary sequences, we have a list per
// channel which contains a bitmap. Each offset represents a sequence. This means we only need to apply a permutation
// once, and all binary sequences reflect the new order. This is a optimized subsumption check, all other operations
// are in return much slower.
//
// Status: failure. After a permutation is applied the sequences are no longer in sorted order, meaning we can't
//
//	do a proper subsumption check unless we extract each sequence and then compare...
type Warhol struct {
	Channels []bitmap.Bitmap
	Mask     sortnet.BinarySequence
	*sortnet.SetMetadata
}

func (s *Warhol) Metadata() *sortnet.SetMetadata {
	return s.SetMetadata
}

func (s *Warhol) Contains(seq sortnet.BinarySequence) bool {
	for it := sortnet.NewSequenceIterator(seq); !it.Empty(); {
		channel := it.Next()
		if !s.Channels[channel].Contains(uint32(seq)) {
			return false
		}
	}

	return true
}

func (s *Warhol) ContainsInPartition(seq sortnet.BinarySequence, _ int) bool {
	return s.Contains(seq)
}

func (s *Warhol) Add(seq sortnet.BinarySequence) {
	if !s.Contains(seq) {
		partition := seq.OnesCount()
		s.SetMetadata.Add(seq, partition)

		for it := sortnet.NewSequenceIterator(seq); !it.Empty(); {
			channel := it.Next()
			s.Channels[channel].Set(uint32(seq))
		}
	}

	s.Mask |= seq
}

func (s *Warhol) Derive(network sortnet.Network) sortnet.OutputSet {
	output := NewEmptyWarhol(len(s.Channels))

	mask := sortnet.SequenceMask(len(s.Channels))
	for target := sortnet.BinarySequence(1); target < mask; target++ {
		if s.Contains(target) {
			output.Add(network.Transform(target))
		}
	}
	return output
}

func (s *Warhol) Size() int {
	return s.SetMetadata.Size
}

func (s *Warhol) PartitionSize(p int) int {
	return s.SetMetadata.PartitionSizes[p]
}

func bitmapSubsumes(left, right bitmap.Bitmap) bool {
	for i := 0; i < len(left); i++ {
		if (left[i] & right[i]) != left[i] {
			return false
		}
	}

	return true
}

func (s *Warhol) IsSubset(otherA sortnet.OutputSet, permutation sortnet.PermutationMap) bool {
	other := otherA.(*Warhol)

	for channel := range s.Channels {
		target := channel
		if permutation != nil {
			target = permutation[channel]
		}

		if !bitmapSubsumes(s.Channels[channel], other.Channels[target]) {
			return false
		}
	}
	return true
}
