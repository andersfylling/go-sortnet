package main

import (
	"context"
	"github.com/andersfylling/go-sortnet/sortnet"
	"github.com/kyroy/kdtree"
	"golang.org/x/sync/errgroup"
)

type Container struct {
	sets  []sortnet.OutputSet
	trees map[int]*KDTree
}

func (c *Container) Insert(sets []sortnet.OutputSet) {
	if c.trees == nil {
		c.trees = map[int]*KDTree{}
	}

	for _, set := range sets {
		if set == nil {
			continue
		}

		size := set.Size()
		if _, ok := c.trees[size]; !ok {
			c.trees[size] = NewKDTree()
		}

		point := CreateMetadataPoint(set)
		c.trees[size].Insert([]*Metadata{point})
		c.sets = append(c.sets, set)
	}
}

func (c *Container) Balance() {
	for _, tree := range c.trees {
		tree.Balance()
	}
}

func (c *Container) Search(set sortnet.OutputSet, result chan<- *Work) {
	point := CreateMetadataPoint(set)
	size := set.Size()

	g, _ := errgroup.WithContext(context.Background())
	work := make(chan int)

	g.Go(func() error {
		defer close(work)
		for treeSize := range c.trees {
			if treeSize < size {
				continue
			}

			work <- treeSize
		}
		return nil
	})

	for i := 0; i < Workers; i++ {
		g.Go(func() error {
			for treeSize := range work {
				tree := c.trees[treeSize]
				matches := tree.FindCandidates(point)
				for _, id := range matches {
					target := c.sets[int(id)]
					if target == nil || set == target {
						continue
					}

					result <- &Work{
						set: target,
						id:  target.Metadata().NetworkID,
					}
				}
			}
			return nil
		})
	}

	_ = g.Wait()
}

func (c *Container) Prune(ids []sortnet.NetworkID) {
	for _, id := range ids {
		c.sets[int(id)] = nil
	}
}

func (c *Container) Rebalance() {
	for size, tree := range c.trees {
		var remaining []kdtree.Point
		ids, ps := tree.Points()
		for i := range ids {
			if c.sets[ids[i]] != nil {
				remaining = append(remaining, ps[i])
			}
		}

		c.trees[size] = NewKDTree()
		c.trees[size].InsertRaw(remaining)
		c.trees[size].Balance()
	}
}
