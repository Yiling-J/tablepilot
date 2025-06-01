package source

import (
	"context"
	"math/rand/v2"
	"os"

	"github.com/Yiling-J/tablepilot/ent"
	"github.com/Yiling-J/tablepilot/ent/schema"
)

type BasicSource struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type Source interface {
	Next(ctx context.Context, idx int) (*schema.CellValue, error)
	Total() int
}

type Indexer struct {
	source   Source
	column   *ent.TableColumn
	repeated int
	total    int
	current  int
	picked   map[int]bool
	rng      *rand.Rand
}

func NewIndexer(source Source, column *ent.TableColumn) *Indexer {
	if column.Repeat == 0 {
		column.Repeat = 1
	}
	indexer := &Indexer{
		source:  source,
		column:  column,
		total:   source.Total(),
		current: -1,
		picked:  map[int]bool{},
	}
	// To ensure a replayable snapshot, we need a fixed random sequence in both record step and test step.
	if v, _ := os.LookupEnv("TABLEPILOT_SNAPSHOT_RECORD"); len(v) > 0 {
		indexer.rng = rand.New(rand.NewPCG(90, 723))
	}
	if v, _ := os.LookupEnv("TABLEPILOT_SNAPSHOT_TEST"); len(v) > 0 {
		indexer.rng = rand.New(rand.NewPCG(90, 723))
	}
	return indexer
}

func (i *Indexer) nextIndex() int {
	if i.column.Repeat > 1 && i.repeated < i.column.Repeat && i.current != -1 {
		i.repeated += 1
		return i.current
	}
	i.repeated = 1

	if i.column.Random {
		if !i.column.Replacement {
			options := []int{}
			if len(i.picked) == i.total {
				i.picked = map[int]bool{}
			}
			for j := range i.total {
				if _, ok := i.picked[j]; !ok {
					options = append(options, j)
				}
			}
			if i.rng != nil {
				i.current = options[i.rng.IntN(len(options))]
			} else {
				i.current = options[rand.IntN(len(options))]
			}
			i.picked[i.current] = true
		} else {
			if i.rng != nil {
				i.current = i.rng.IntN(i.total)
			} else {
				i.current = rand.IntN(i.total)
			}
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
	switch ts := i.source.(type) {
	case *LinkedSource:
		return ts.NextLinked(ctx, i.nextIndex(), i.column.LinkedColumn, i.column.LinkedContextColumns)
	case *CsvSource:
		return ts.NextLinked(ctx, i.nextIndex(), i.column.LinkedColumn, i.column.LinkedContextColumns)
	case *ParquetSource:
		return ts.NextLinked(ctx, i.nextIndex(), i.column.LinkedColumn, i.column.LinkedContextColumns)
	default:
		return i.source.Next(ctx, i.nextIndex())
	}
}
