package source

import (
	"context"
	"math/rand/v2"

	"github.com/Yiling-J/tablepilot/ent/schema"
)

type Source interface {
	Next(ctx context.Context, idx int) (*schema.CellValue, error)
	Total() int
}

type Indexer struct {
	source      Source
	Random      bool `json:"random,omitempty"`
	Replacement bool `json:"replacement,omitempty"`
	Repeat      int  `json:"repeat,omitempty"`
	repeated    int
	total       int
	current     int
	picked      map[int]bool
}

func NewIndexer(source Source, random, replacement bool, repeat int) *Indexer {
	if repeat == 0 {
		repeat = 1
	}
	return &Indexer{
		source:      source,
		Random:      random,
		Replacement: replacement,
		Repeat:      repeat,
		total:       source.Total(),
		current:     -1,
		picked:      map[int]bool{},
	}
}

func (i *Indexer) nextIndex() int {
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

func (i *Indexer) Next(ctx context.Context) (*schema.CellValue, error) {
	return i.source.Next(ctx, i.nextIndex())
}
