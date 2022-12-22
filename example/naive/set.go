package main

import "github.com/andersfylling/go-sortnet/sortnet"

func NewSliceOutputSet(sequenceSize int) *SliceOutputSet {
	return sortnet.PopulateOutputSet(&SliceOutputSet{}, sequenceSize).(*SliceOutputSet)
}

type SliceOutputSet struct {
	Sequences []sortnet.BinarySequence
}

func (s *SliceOutputSet) Contains(seq sortnet.BinarySequence) bool {
	for i := range s.Sequences {
		if seq == s.Sequences[i] {
			return true
		}
	}
	return false
}

func (s *SliceOutputSet) Add(seq sortnet.BinarySequence) {
	if !s.Contains(seq) {
		s.Sequences = append(s.Sequences, seq)
	}
}

func (s *SliceOutputSet) Derive(network sortnet.Network) *SliceOutputSet {
	output := &SliceOutputSet{}

	for i := range s.Sequences {
		output.Add(network.Transform(s.Sequences[i]))
	}
	return output
}

func (s *SliceOutputSet) Size() int {
	return len(s.Sequences)
}

func (s *SliceOutputSet) IsSubset(other *SliceOutputSet) bool {
	for i := range s.Sequences {
		if !other.Contains(s.Sequences[i]) {
			return false
		}
	}
	return true
}
