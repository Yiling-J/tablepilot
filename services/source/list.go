package source

import (
	"bufio"
	"context"
	"os"

	"github.com/Yiling-J/tablepilot/ent/schema"
)

type ListSource struct {
	BasicSource
	Options []string `json:"options"`
	File    string   `json:"file,omitempty"`
}

func (ls *ListSource) Init(ctx context.Context, dir string) error {
	if ls.File != "" {
		root, err := os.OpenRoot(dir)
		if err != nil {
			return err
		}

		file, err := root.FS().Open(ls.File)
		if err != nil {
			return err
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				ls.Options = append(ls.Options, scanner.Text())
			}
		}
		if err := scanner.Err(); err != nil {
			return err
		}
	}
	return nil
}

func (ls *ListSource) Next(ctx context.Context, idx int) (*schema.CellValue, error) {
	return &schema.CellValue{Value: ls.Options[idx]}, nil
}

func (ls *ListSource) Total() int {
	return len(ls.Options)
}
