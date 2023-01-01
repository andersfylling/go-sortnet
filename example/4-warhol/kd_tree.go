package main

import (
	"github.com/andersfylling/go-sortnet/sortnet"
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
	ID             sortnet.NetworkID
}

func (md *Metadata) Coordinates() []float64 {
	coordinates := []float64{}
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

func (kdt *KDTree) FindCandidates(point *Metadata) []sortnet.NetworkID {
	var subsumptionRange []float64
	for _, low := range point.Coordinates() {
		subsumptionRange = append(subsumptionRange, low)
		subsumptionRange = append(subsumptionRange, math.MaxFloat64)
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