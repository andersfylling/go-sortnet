package outputset

import "github.com/andersfylling/go-sortnet/sortnet"

func NewEmptyPartitionedOrdered() *PartitionedOrdered {
	set := &PartitionedOrdered{
		Metadata: &Metadata{},
	}
	return set
}

func NewPartitionedOrdered(sequenceSize int) *PartitionedOrdered {
	return sortnet.PopulateOutputSet(NewEmptyPartitionedOrdered(), sequenceSize).(*PartitionedOrdered)
}

type PartitionedOrdered struct {
	Partitions []map[sortnet.BinarySequence]struct{}
	Sequences  [][]sortnet.BinarySequence
	*Metadata
}

func (s *PartitionedOrdered) Contains(seq sortnet.BinarySequence) bool {
	partition := seq.OnesCount()
	return s.ContainsInPartition(seq, partition)
}

func (s *PartitionedOrdered) ContainsInPartition(seq sortnet.BinarySequence, pi int) bool {
	// disadvantage: "runtime.mapaccess" does hashing
	_, ok := s.Partitions[pi][seq]
	return ok
}

func (s *PartitionedOrdered) Add(seq sortnet.BinarySequence) {
	partition := seq.OnesCount()
	if partition >= len(s.Partitions) {
		for i := len(s.Partitions); i <= partition; i++ {
			s.Partitions = append(s.Partitions, map[sortnet.BinarySequence]struct{}{})
			s.Sequences = append(s.Sequences, []sortnet.BinarySequence{})
			s.OnesMasks = append(s.OnesMasks, 0)
			s.ZerosMasks = append(s.ZerosMasks, 0)
		}
	}

	if !s.ContainsInPartition(seq, partition) {
		s.Sequences[partition] = append(s.Sequences[partition], seq)
	}
	s.Partitions[partition][seq] = empty
	s.OnesMasks[partition] |= seq
	s.ZerosMasks[partition] |= ^seq
}

func (s *PartitionedOrdered) Derive(network sortnet.Network) *PartitionedOrdered {
	output := NewEmptyPartitionedOrdered()

	for _, partition := range s.Sequences {
		for _, seq := range partition {
			output.Add(network.Transform(seq))
		}
	}
	return output
}

func (s *PartitionedOrdered) Size() int {
	var size int
	for _, partition := range s.Partitions {
		size += len(partition)
	}

	return size
}

func (s *PartitionedOrdered) PartitionSize(p int) int {
	return len(s.Partitions[p])
}

func (s *PartitionedOrdered) IsSubset(other *PartitionedOrdered, permutation sortnet.PermutationMap) bool {
	for pi := range s.Sequences {
		for _, seq := range s.Sequences[pi] {
			if permutation != nil {
				seq = sortnet.ApplyPermutation(seq, permutation)
			}
			if !other.ContainsInPartition(seq, pi) {
				return false
			}
		}
	}
	return true
}
