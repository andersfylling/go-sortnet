package outputset

import "github.com/andersfylling/go-sortnet/sortnet"

var empty = struct{}{}

type NewSet = func(channels int) sortnet.OutputSet
