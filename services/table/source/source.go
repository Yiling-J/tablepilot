package source

import (
	"context"
	"math/rand/v2"
	"tablepilot/ent/schema"
)

type Source interface {
	Next(ctx context.Context) (*schema.CellValue, error)
}

type indexer struct {
	Random      bool `json:"random,omitempty"`
	Replacement bool `json:"replacement,omitempty"`
	Repeat      int  `json:"repeat,omitempty"`
	repeated    int
	total       int
	current     int
	picked      map[int]bool
}

func newIndexer(random, replacement bool, total, repeat int) indexer {
	if repeat == 0 {
		repeat = 1
	}
	return indexer{
		Random:      random,
		Replacement: replacement,
		Repeat:      repeat,
		total:       total,
		current:     -1,
		picked:      map[int]bool{},
	}
}

func (i *indexer) nextIndex() int {
	if i.Repeat > 1 && i.repeated < i.Repeat && i.current != -1 {
		i.repeated += 1
		return i.current
	}
	i.repeated = 1

	if i.Random {
		if !i.Replacement {
			options := []int{}
			if len(i.picked) == i.total {
				i.picked = map[int]bool{}
			}
			for j := 0; j < i.total; j++ {
				if _, ok := i.picked[j]; !ok {
					options = append(options, j)
				}
			}
			i.current = options[rand.IntN(len(options))]
			i.picked[i.current] = true
		} else {
			i.current = rand.IntN(i.total)
		}
	} else {
		i.current += 1
		if i.current == i.total {
			i.current = 0
		}
	}
	return i.current
}
