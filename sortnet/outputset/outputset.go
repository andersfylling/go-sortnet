package outputset

import "github.com/andersfylling/go-sortnet/sortnet"

var empty = struct{}{}

type Metadata struct {
	OnesMasks  []sortnet.BinarySequence
	ZerosMasks []sortnet.BinarySequence
}
