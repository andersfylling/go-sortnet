package main

import "github.com/andersfylling/go-sortnet/sortnet"

func NewSet(sequenceSize int) *Set {
	return sortnet.PopulateOutputSet(&Set{}, sequenceSize).(*Set)
}

type Set struct {
	Partitions [][]sortnet.BinarySequence
	OnesMasks  []sortnet.BinarySequence
	ZerosMasks []sortnet.BinarySequence
}

func (s *Set) Contains(seq sortnet.BinarySequence) bool {
	partition := seq.OnesCount()
	return s.ContainsInPartition(seq, partition)
}

func (s *Set) ContainsInPartition(seq sortnet.BinarySequence, partition int) bool {
	for _, storedSeq := range s.Partitions[partition] {
		if seq == storedSeq {
			return true
		}
	}
	return false
}

func (s *Set) Add(seq sortnet.BinarySequence) {
	partition := seq.OnesCount()
	if partition >= len(s.Partitions) {
		for i := len(s.Partitions); i <= partition; i++ {
			s.Partitions = append(s.Partitions, []sortnet.BinarySequence{})
			s.OnesMasks = append(s.OnesMasks, 0)
			s.ZerosMasks = append(s.ZerosMasks, 0)
		}
	}
	if !s.ContainsInPartition(seq, partition) {
		s.Partitions[partition] = append(s.Partitions[partition], seq)
		s.OnesMasks[partition] |= seq
		s.ZerosMasks[partition] |= ^seq
	}
}

func (s *Set) Derive(network sortnet.Network) *Set {
	output := &Set{}

	for _, partition := range s.Partitions {
		for _, seq := range partition {
			output.Add(network.Transform(seq))
		}
	}
	return output
}

func (s *Set) Size() int {
	var size int
	for _, partition := range s.Partitions {
		size += len(partition)
	}

	return size
}

func (s *Set) PartitionSize(p int) int {
	return len(s.Partitions[p])
}

func (s *Set) IsSubset(other *Set) bool {
	for p := range s.Partitions {
		for _, seq := range s.Partitions[p] {
			if !other.ContainsInPartition(seq, p) {
				return false
			}
		}
	}
	return true
}
