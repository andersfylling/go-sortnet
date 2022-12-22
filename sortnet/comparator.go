package sortnet

var comparatorsCache = map[int][]Comparator{}

func AllComparatorCombinations(sequenceSize int) []Comparator {
	if _, ok := comparatorsCache[sequenceSize]; !ok {
		var comparators []Comparator
		for offset1 := sequenceSize - 1; offset1 > 0; offset1-- {
			for offset2 := offset1 - 1; offset2 >= 0; offset2-- {
				comparators = append(comparators, Comparator{
					From: offset1,
					To:   offset2,
				})
			}
		}

		comparatorsCache[sequenceSize] = comparators
	}

	return comparatorsCache[sequenceSize]
}

type Comparator struct {
	From int
	To   int
}
