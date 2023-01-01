package outputset

import "github.com/andersfylling/go-sortnet/sortnet"

func NewEmptyPartitionedUnordered() sortnet.OutputSet {
	set := &PartitionedUnordered{
		SetMetadata: &sortnet.SetMetadata{},
	}
	return set
}

func NewPartitionedUnordered(channels int) sortnet.OutputSet {
	return sortnet.PopulateOutputSet(NewEmptyPartitionedUnordered(), channels)
}

type PartitionedUnordered struct {
	Partitions [][]sortnet.BinarySequence
	*sortnet.SetMetadata
}

func (s *PartitionedUnordered) Metadata() *sortnet.SetMetadata {
	return s.SetMetadata
}

func (s *PartitionedUnordered) Contains(seq sortnet.BinarySequence) bool {
	partition := seq.OnesCount()
	return s.ContainsInPartition(seq, partition)
}

func (s *PartitionedUnordered) ContainsInPartition(seq sortnet.BinarySequence, partition int) bool {
	for _, storedSeq := range s.Partitions[partition] {
		if seq == storedSeq {
			return true
		}
	}
	return false
}

func (s *PartitionedUnordered) Add(seq sortnet.BinarySequence) {
	partition := seq.OnesCount()
	if partition >= len(s.Partitions) {
		for i := len(s.Partitions); i <= partition; i++ {
			s.Partitions = append(s.Partitions, []sortnet.BinarySequence{})
		}
	}
	if !s.ContainsInPartition(seq, partition) {
		s.Partitions[partition] = append(s.Partitions[partition], seq)
		s.SetMetadata.Add(seq, partition)
	}
}

func (s *PartitionedUnordered) Derive(network sortnet.Network) sortnet.OutputSet {
	output := NewEmptyPartitionedUnordered()

	for _, partition := range s.Partitions {
		for _, seq := range partition {
			output.Add(network.Transform(seq))
		}
	}
	return output
}

func (s *PartitionedUnordered) Size() int {
	if s.SetMetadata.Size == 0 {
		var size int
		for _, partition := range s.Partitions {
			size += len(partition)
		}

		s.SetMetadata.Size = size
	}
	return s.SetMetadata.Size
}

func (s *PartitionedUnordered) PartitionSize(p int) int {
	return len(s.Partitions[p])
}

func (s *PartitionedUnordered) IsSubset(other sortnet.OutputSet, permutationMap sortnet.PermutationMap) bool {
	for p := range s.Partitions {
		for _, seq := range s.Partitions[p] {
			if permutationMap != nil {
				seq = sortnet.ApplyPermutation(seq, permutationMap)
			}
			if !other.ContainsInPartition(seq, p) {
				return false
			}
		}
	}
	return true
}
