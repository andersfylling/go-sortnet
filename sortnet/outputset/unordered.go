package outputset

import "github.com/andersfylling/go-sortnet/sortnet"

func NewUnordered(sequenceSize int) *Unordered {
	return sortnet.PopulateOutputSet(&Unordered{}, sequenceSize).(*Unordered)
}

type Unordered struct {
	Sequences []sortnet.BinarySequence
}

func (s *Unordered) Contains(seq sortnet.BinarySequence) bool {
	for i := range s.Sequences {
		if seq == s.Sequences[i] {
			return true
		}
	}
	return false
}

func (s *Unordered) Add(seq sortnet.BinarySequence) {
	if !s.Contains(seq) {
		s.Sequences = append(s.Sequences, seq)
	}
}

func (s *Unordered) Derive(network sortnet.Network) *Unordered {
	output := &Unordered{}

	for i := range s.Sequences {
		output.Add(network.Transform(s.Sequences[i]))
	}
	return output
}

func (s *Unordered) Size() int {
	return len(s.Sequences)
}

func (s *Unordered) IsSubset(other *Unordered) bool {
	for i := range s.Sequences {
		if !other.Contains(s.Sequences[i]) {
			return false
		}
	}
	return true
}
