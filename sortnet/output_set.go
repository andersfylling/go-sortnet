package sortnet

type OutputSet interface {
	Derive(Network) OutputSet
	Add(sequence BinarySequence)
	Size() int
	IsSubset(OutputSet, PermutationMap) bool
	Contains(BinarySequence) bool
	ContainsInPartition(seq BinarySequence, partition int) bool
	Metadata() *SetMetadata
}

func PopulateOutputSet(set OutputSet, channels int) OutputSet {
	mask := SequenceMask(channels)
	for seq := BinarySequence(1); seq < mask; seq++ {
		set.Add(seq)
	}

	return set
}

type SetMetadata struct {
	Size           int
	PartitionSizes []int
	OnesMasks      []BinarySequence
	ZerosMasks     []BinarySequence
}

func (md *SetMetadata) Add(seq BinarySequence, partition int) {
	if partition >= len(md.PartitionSizes) {
		for i := len(md.PartitionSizes); i <= partition; i++ {
			md.OnesMasks = append(md.OnesMasks, 0)
			md.ZerosMasks = append(md.ZerosMasks, 0)
			md.PartitionSizes = append(md.PartitionSizes, 0)
		}
	}

	md.Size++
	md.PartitionSizes[partition]++
	md.OnesMasks[partition] |= seq
	md.ZerosMasks[partition] |= ^seq
}

// ST1 check total size of the output
func (md *SetMetadata) ST1(other *SetMetadata) bool {
	return md.Size <= other.Size
}

// ST2 check the size of corresponding clusters/partitions
func (md *SetMetadata) ST2(other *SetMetadata) bool {
	for pi := 0; pi < len(md.PartitionSizes); pi++ {
		if md.PartitionSizes[pi] > other.PartitionSizes[pi] {
			return false
		}
	}

	return true
}

// ST3 check the ones and the zeros
func (md *SetMetadata) ST3(other *SetMetadata) bool {
	for pi := 0; pi < len(md.PartitionSizes); pi++ {
		if md.OnesMasks[pi].OnesCount() > other.OnesMasks[pi].OnesCount() {
			return false
		}
	}
	for pi := 0; pi < len(md.PartitionSizes); pi++ {
		if md.ZerosMasks[pi].OnesCount() > other.ZerosMasks[pi].OnesCount() {
			return false
		}
	}

	return true
}

// ST4 check all permutations
func (md *SetMetadata) ST4(other *SetMetadata) bool {
	panic("ST4 is instead implemented as a permutation generator - see sortnet/permutation.go")
}
