package sortnet

import (
	"fmt"
	"strings"
)

type Network interface {
	Transform(BinarySequence) BinarySequence
	Derive(comparators []Comparator) []Network
}

type ComparatorNetwork struct {
	comparators []Comparator
}

func (n *ComparatorNetwork) String() string {
	// simple ascii representation
	// each dot represents a node in the channel
	// each row represents a single channel

	var channels int
	for _, comparator := range n.comparators {
		if comparator.From > channels {
			channels = comparator.From
		} else if comparator.To > channels {
			channels = comparator.To
		}
	}

	networkStr := strings.Builder{}
	for channel := 0; channel <= channels; channel++ {
		channelStr := strings.Builder{}
		channelStr.WriteString(fmt.Sprintf("%d: ", channel))
		for _, comparator := range n.comparators {
			r := "-"
			if comparator.To == channel {
				r = "+"
			} else if comparator.From == channel {
				r = "^"
			} else if comparator.To < channel && comparator.From > channel {
				r = "|"
			}

			channelStr.WriteString(fmt.Sprintf(" %s", r))
		}
		networkStr.WriteString(channelStr.String() + "\n")
	}

	return networkStr.String()
}

func (n *ComparatorNetwork) Transform(seq BinarySequence) BinarySequence {
	for _, comparator := range n.comparators {
		leftMostBitMask := BinarySequence(1 << comparator.From)
		rightMostBitMask := BinarySequence(1 << comparator.To)
		mask := leftMostBitMask | rightMostBitMask

		if mask&seq == leftMostBitMask {
			seq = (seq ^ leftMostBitMask) | rightMostBitMask
		}
	}

	return seq
}

func (n *ComparatorNetwork) Derive(comparators []Comparator) []Network {
	var children []Network

	for _, comparator := range comparators {
		if len(n.comparators) > 0 && comparator == n.comparators[len(n.comparators)-1] {
			continue
		}

		child := &ComparatorNetwork{
			comparators: make([]Comparator, len(n.comparators)),
		}

		copy(child.comparators, n.comparators)
		child.comparators = append(child.comparators, comparator)

		children = append(children, child)
	}

	return children
}
