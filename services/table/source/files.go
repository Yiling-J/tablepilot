package source

import (
	"context"
	"os"

	"github.com/Yiling-J/tablepilot/ent/schema"
	"go.uber.org/zap"
)

type FilesSource struct {
	Type  string   `json:"type"`
	Paths []string `json:"paths"`
	files []string
}

func (f *FilesSource) getRoot(ctx context.Context, logger *zap.SugaredLogger, dir string) (*os.Root, string, error) {
	var root *os.Root
	rootPath := dir
	var err error
	root, err = os.OpenRoot(dir)
	if err != nil {
		return nil, "", err
	}
	return root, rootPath, nil
}

func (f *FilesSource) Init(ctx context.Context, logger *zap.SugaredLogger, dir string) error {
	root, _, err := f.getRoot(ctx, logger, dir)
	if err != nil {
		return err
	}
	fileSystem := root.FS()
	f.files, err = parsePaths(fileSystem, f.Paths)
	if err != nil {
		return err
	}
	return nil
}

func (f *FilesSource) Next(ctx context.Context, idx int) (*schema.CellValue, error) {
	return &schema.CellValue{Value: f.files[idx]}, nil
}

func (f *FilesSource) Total() int {
	return len(f.files)
}
