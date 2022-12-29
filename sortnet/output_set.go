package sortnet

type OutputSet interface {
	Add(sequence BinarySequence)
}

func PopulateOutputSet(set OutputSet, sequenceSize int) OutputSet {
	mask := SequenceMask(sequenceSize)
	for seq := BinarySequence(0); seq < mask; seq++ {
		set.Add(seq)
	}

	return set
}
