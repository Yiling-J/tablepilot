package csvindexer

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"

	"github.com/spf13/cast"
)

const chunkSize = 50

type FileOffset struct {
	File   uint16
	Total  uint8 // each chunk has max 50 rows
	Offset int64
}

type CSVIndexer struct {
	files     []string
	total     int
	positions []FileOffset
	columns   []string
}

func NewCSVIndexer(files []string) (*CSVIndexer, error) {
	var positions []FileOffset
	var total int
	var columns []string

	for i, file := range files {
		f, err := os.Open(file)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		r := csv.NewReader(f)

		fileColumns, err := r.Read()
		if err != nil {
			return nil, fmt.Errorf("error reading header from file %s: %w", file, err)
		}

		if columns == nil {
			columns = fileColumns
		} else if !reflect.DeepEqual(columns, fileColumns) {
			return nil, fmt.Errorf("headers in file %s do not match previous files. Expected: %v, Got: %v", file, columns, fileColumns)
		}
		currentPosition := FileOffset{File: cast.ToUint16(i), Offset: r.InputOffset()}
		for {
			_, err := r.Read()
			if err == io.EOF {
				if total%chunkSize != 0 {
					positions = append(positions, currentPosition)
				}
				break
			} else if err != nil {
				return nil, err
			}
			total++
			currentPosition.Total++
			// current chunk end
			if currentPosition.Total%chunkSize == 0 {
				positions = append(positions, currentPosition)
				currentPosition = FileOffset{File: cast.ToUint16(i), Offset: r.InputOffset()}
			}
		}
	}

	return &CSVIndexer{
		files:     files,
		total:     total,
		positions: positions,
		columns:   columns,
	}, nil
}

func (r *CSVIndexer) Fetch(idx int) ([]string, error) {
	start := 0
	var pos *FileOffset
	for _, p := range r.positions {
		if idx >= start && idx < start+int(p.Total) {
			pos = &p
			break
		}
		start += int(p.Total)
	}
	if pos == nil {
		return nil, errors.New("position for index not found")
	}
	file := r.files[pos.File]
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	_, err = f.Seek(pos.Offset, io.SeekStart)
	if err != nil {
		return nil, err
	}
	reader := csv.NewReader(f)
	nth := idx % chunkSize
	for i := 0; ; i++ {
		if i >= r.total {
			break
		}
		record, err := reader.Read()
		if err != nil {
			return nil, err
		}
		if i == nth {
			return record, nil
		}
	}
	return nil, errors.New("index not found")
}

func (r *CSVIndexer) Total() int {
	return r.total
}

func (r *CSVIndexer) Columns() []string {
	return r.columns
}
