package outputset

import "github.com/andersfylling/go-sortnet/sortnet"

func NewEmptyUnordered() sortnet.OutputSet {
	set := &Unordered{
		SetMetadata: &sortnet.SetMetadata{},
	}
	return set
}

func NewUnordered(channels int) sortnet.OutputSet {
	return sortnet.PopulateOutputSet(NewEmptyUnordered(), channels)
}

type Unordered struct {
	Sequences []sortnet.BinarySequence
	*sortnet.SetMetadata
}

func (s *Unordered) Metadata() *sortnet.SetMetadata {
	return s.SetMetadata
}

func (s *Unordered) Contains(seq sortnet.BinarySequence) bool {
	for i := range s.Sequences {
		if seq == s.Sequences[i] {
			return true
		}
	}
	return false
}

func (s *Unordered) ContainsInPartition(seq sortnet.BinarySequence, _ int) bool {
	return s.Contains(seq)
}

func (s *Unordered) Add(seq sortnet.BinarySequence) {
	if !s.Contains(seq) {
		s.Sequences = append(s.Sequences, seq)
		s.SetMetadata.Add(seq, seq.OnesCount())
	}
}

func (s *Unordered) Derive(network sortnet.Network) sortnet.OutputSet {
	output := NewEmptyUnordered()

	for i := range s.Sequences {
		output.Add(network.Transform(s.Sequences[i]))
	}
	return output
}

func (s *Unordered) Size() int {
	return s.SetMetadata.Size
}

func (s *Unordered) IsSubset(other sortnet.OutputSet, permutationMap sortnet.PermutationMap) bool {
	for _, seq := range s.Sequences {
		if permutationMap != nil {
			seq = sortnet.ApplyPermutation(seq, permutationMap)
		}
		if !other.Contains(seq) {
			return false
		}
	}
	return true
}
