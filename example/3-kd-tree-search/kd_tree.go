package main

import (
	"github.com/kyroy/kdtree"
	"github.com/kyroy/kdtree/kdrange"
	"github.com/kyroy/kdtree/points"
	"math"
)

type Metadata struct {
	Size           int
	Ones           []int
	Zeros          []int
	PartitionSizes []int
	Index          int
}

func (md *Metadata) Coordinates() []float64 {
	coordinates := []float64{
		float64(md.Size),
	}
	for pi := range md.PartitionSizes {
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
	for i := range sets {
		kdt.tree.Insert(points.NewPoint(sets[i].Coordinates(), sets[i].Index))
	}
}

func (kdt *KDTree) Balance() {
	kdt.tree.Balance()
}

func (kdt *KDTree) FindCandidates(point *Metadata) []int {
	var subsumptionRange []float64
	for _, low := range point.Coordinates() {
		subsumptionRange = append(subsumptionRange, low)
		subsumptionRange = append(subsumptionRange, math.MaxFloat64)
	}
	filter := kdrange.New(subsumptionRange...)
	matches := kdt.tree.RangeSearch(filter)

	var indexes []int
	for i := range matches {
		match := matches[i].(*points.Point)
		indexes = append(indexes, match.Data.(int))
	}

	return indexes
}
