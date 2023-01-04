package main

import (
	"github.com/andersfylling/go-sortnet/sortnet"
	"github.com/kyroy/kdtree"
	"github.com/kyroy/kdtree/kdrange"
	"github.com/kyroy/kdtree/points"
	"math"
)

type SearchDirection int

const (
	DirectionSubset SearchDirection = iota
	DirectionEqual
	DirectionEqualAndSuperset
	DirectionSuperset
)

type Metadata struct {
	Size           int
	Ones           []int
	Zeros          []int
	PartitionSizes []int
	ID             sortnet.NetworkID
}

func (md *Metadata) Coordinates() []float64 {
	coordinates := make([]float64, 0, len(md.PartitionSizes)+len(md.Ones)+len(md.Zeros))
	for pi := 0; pi < len(md.PartitionSizes); pi++ {
		coordinates = append(coordinates, float64(md.PartitionSizes[pi]))
		coordinates = append(coordinates, float64(md.Ones[pi]))
		coordinates = append(coordinates, float64(md.Zeros[pi]))
	}

	return coordinates
}

func NewKDTree() *KDTree {
	return &KDTree{
		tree: kdtree.New(nil),
	}
}

type KDTree struct {
	tree *kdtree.KDTree
}

func (kdt *KDTree) Insert(sets []*Metadata) {
	ps := make([]kdtree.Point, 0, len(sets))
	for i := range sets {
		ps = append(ps, points.NewPoint(sets[i].Coordinates(), sets[i].ID))
	}

	kdt.InsertRaw(ps)
}

func (kdt *KDTree) InsertRaw(ps []kdtree.Point) {
	for _, p := range ps {
		kdt.tree.Insert(p)
	}
}

func (kdt *KDTree) Balance() {
	kdt.tree.Balance()
}

func (kdt *KDTree) Points() ([]sortnet.NetworkID, []kdtree.Point) {
	ps := kdt.tree.Points()

	var ids []sortnet.NetworkID
	for _, point := range ps {
		match := point.(*points.Point)
		ids = append(ids, match.Data.(sortnet.NetworkID))
	}

	return ids, ps
}

func (kdt *KDTree) FindCandidates(point *Metadata, direction SearchDirection) []sortnet.NetworkID {
	coordinates := point.Coordinates()
	subsumptionRange := make([]float64, 0, len(coordinates)*2)
	for i := 0; i < len(coordinates); i++ {
		self := coordinates[i]
		switch direction {
		case DirectionSubset:
			subsumptionRange = append(subsumptionRange, 0)
			subsumptionRange = append(subsumptionRange, self)
		case DirectionEqual:
			subsumptionRange = append(subsumptionRange, self)
			subsumptionRange = append(subsumptionRange, self)
		case DirectionEqualAndSuperset:
			subsumptionRange = append(subsumptionRange, self)
			subsumptionRange = append(subsumptionRange, math.MaxFloat64)
		case DirectionSuperset:
			subsumptionRange = append(subsumptionRange, self+1)
			subsumptionRange = append(subsumptionRange, math.MaxFloat64)
		}
	}

	filter := kdrange.New(subsumptionRange...)
	matches := kdt.tree.RangeSearch(filter)

	var indexes []sortnet.NetworkID
	for i := range matches {
		match := matches[i].(*points.Point)
		indexes = append(indexes, match.Data.(sortnet.NetworkID))
	}

	return indexes
}
