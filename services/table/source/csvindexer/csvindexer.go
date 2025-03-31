package csvindexer

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"io/fs"
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
	fs        fs.FS
	files     []string
	total     int
	positions []FileOffset
	columns   []string
}

func NewCSVIndexer(fs fs.FS, files []string) (*CSVIndexer, error) {
	var positions []FileOffset
	var total int
	var columns []string

	for i, file := range files {
		f, err := fs.Open(file)
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
			if errors.Is(err, io.EOF) {
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
		fs:        fs,
		files:     files,
		total:     total,
		positions: positions,
		columns:   columns,
	}, nil
}

func (r *CSVIndexer) Fetch(idx int) ([]string, error) {
	start := 0
	var pos *FileOffset
	// the offset position relative to the chunk, used to get row data from chunk
	offset := 0
	for _, p := range r.positions {
		if idx >= start && idx < start+int(p.Total) {
			pos = &p
			offset = idx - start
			break
		}
		start += int(p.Total)
	}
	if pos == nil {
		return nil, errors.New("position for index not found")
	}
	file := r.files[pos.File]
	f, err := r.fs.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if rs, ok := f.(io.ReadSeeker); ok {
		_, err = rs.Seek(pos.Offset, io.SeekStart)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("file not seekable")
	}
	reader := csv.NewReader(f)
	for i := 0; ; i++ {
		if i >= r.total {
			break
		}
		record, err := reader.Read()
		if err != nil {
			return nil, err
		}
		if i == offset {
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

func (r *CSVIndexer) Range(fn func(row []string) bool) error {
	for _, file := range r.files {
		f, err := r.fs.Open(file)
		if err != nil {
			return err
		}
		reader := csv.NewReader(f)
		// skip header
		_, err = reader.Read()
		if err != nil {
			return err
		}
		for {
			record, err := reader.Read()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				return err
			}
			if !fn(record) {
				break
			}
		}
	}
	return nil
}

func GetColumnsFromFiles(fs fs.FS, files []string) ([]string, error) {
	var columns []string

	for _, file := range files {
		f, err := fs.Open(file)
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
	}

	return columns, nil
}
