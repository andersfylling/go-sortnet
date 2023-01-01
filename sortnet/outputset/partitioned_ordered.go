package outputset

import "github.com/andersfylling/go-sortnet/sortnet"

func NewEmptyPartitionedOrdered() *PartitionedOrdered {
	set := &PartitionedOrdered{
		SetMetadata: &sortnet.SetMetadata{},
	}
	return set
}

func NewPartitionedOrdered(channels int) sortnet.OutputSet {
	return sortnet.PopulateOutputSet(NewEmptyPartitionedOrdered(), channels)
}

type PartitionedOrdered struct {
	Partitions []map[sortnet.BinarySequence]struct{}
	Sequences  [][]sortnet.BinarySequence
	*sortnet.SetMetadata
}

func (s *PartitionedOrdered) Metadata() *sortnet.SetMetadata {
	return s.SetMetadata
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
		}
	}

	if !s.ContainsInPartition(seq, partition) {
		s.Sequences[partition] = append(s.Sequences[partition], seq)
		s.SetMetadata.Add(seq, partition)
	}
	s.Partitions[partition][seq] = empty
}

func (s *PartitionedOrdered) Derive(network sortnet.Network) sortnet.OutputSet {
	output := NewEmptyPartitionedOrdered()

	for _, partition := range s.Sequences {
		for _, seq := range partition {
			output.Add(network.Transform(seq))
		}
	}
	return output
}

func (s *PartitionedOrdered) Size() int {
	return s.SetMetadata.Size
}

func (s *PartitionedOrdered) PartitionSize(p int) int {
	return s.SetMetadata.PartitionSizes[p]
}

func (s *PartitionedOrdered) IsSubset(other sortnet.OutputSet, permutation sortnet.PermutationMap) bool {
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
