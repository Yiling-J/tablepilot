package parquet

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"slices"

	"github.com/parquet-go/parquet-go"
	"golang.org/x/sync/errgroup"
)

const (
	fileReadConcurrency = 32
)

type ParquetGoReader struct {
	fs      fs.FS
	files   []string
	columns []string
	total   int64
}

type groupReader struct {
	reader *parquet.GenericReader[row]
	limit  int64
	offset int64
}

// used by NewGenericRowGroupReader, which require type explicitly
type row any

// pack parquet.File and closer together, so we can close the file after read the content
type pfile struct {
	rc io.Closer
	pf *parquet.File
}

type loader struct {
	groupReaders []groupReader
	files        []io.Closer
}

func (l *loader) close() {
	for _, f := range l.files {
		_ = f.Close()
	}
}

func NewParquetGoReader(ctx context.Context, fs fs.FS, files []string) (*ParquetGoReader, error) {
	pr := &ParquetGoReader{
		fs:    fs,
		files: files,
	}

	pfs, err := pr.loadFiles(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		for _, f := range pfs {
			_ = f.rc.Close()
		}
	}()

	for _, schema := range pfs[0].pf.Metadata().Schema {
		if schema.Type == nil {
			continue
		}
		pr.columns = append(pr.columns, schema.Name)
	}

	for _, f := range pfs {
		pr.total += f.pf.NumRows()
	}
	return pr, nil
}

func (p *ParquetGoReader) build(files []*pfile, limit, offset int64) []groupReader {
	var done bool
	// recalculate required groups
	readers := []groupReader{}
	var processed int64
	for _, f := range files {
		if done {
			break
		}

		for _, rg := range f.pf.RowGroups() {
			if done {
				break
			}

			// the global start/end row index of current group
			start := processed
			end := processed + rg.NumRows()

			if offset >= start && offset <= end {
				// the local offset/limit size of current group
				var groupOffset = offset - processed
				var groupLimit = groupOffset + limit

				// groupLimit > nr: need more data from next group
				if nr := rg.NumRows(); groupLimit > nr {
					groupLimit = nr
					// the global offset should be end of current group
					offset = end
					// the global limit should be remaining rows need to be fetched
					limit -= (nr - groupOffset)
				} else {
					done = true
				}
				readers = append(readers, groupReader{
					reader: parquet.NewGenericRowGroupReader[row](rg),
					limit:  groupLimit,
					offset: groupOffset,
				})
			}
			processed += rg.NumRows()
		}
	}
	return readers
}

func (p *ParquetGoReader) loadFiles(ctx context.Context) ([]*pfile, error) {
	var base int
	files := make([]*pfile, len(p.files))
	for c := range slices.Chunk(p.files, fileReadConcurrency) {
		g, _ := errgroup.WithContext(ctx)
		for i := 0; i < len(c); i++ {
			index := base + i
			g.Go(func() error {
				pf, err := p.parquetFile(p.files[index])
				if err != nil {
					return err
				}
				files[index] = pf
				return nil
			})
		}
		if err := g.Wait(); err != nil {
			return nil, err
		}
		base += len(c)
	}
	return files, nil
}

func (p *ParquetGoReader) loader(ctx context.Context, limit, offset int64) (*loader, error) {
	files, err := p.loadFiles(ctx)
	if err != nil {
		return nil, err
	}

	rc := []io.Closer{}
	for _, f := range files {
		rc = append(rc, f.rc)
	}

	return &loader{
		groupReaders: p.build(files, limit, offset),
		files:        rc,
	}, nil
}

type ReaderAtCloser interface {
	io.ReaderAt
	io.Closer
}

func (p *ParquetGoReader) parquetFile(path string) (*pfile, error) {
	// we need to load file content later, so can't defer close here
	// instead, only close the object manually when there is error
	f, err := p.fs.Open(path)
	if err != nil {
		return nil, err
	}
	s, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if obj, ok := f.(ReaderAtCloser); ok {
		pf, err := parquet.OpenFile(obj, s.Size())
		if err != nil {
			_ = obj.Close()
			return nil, err
		}

		return &pfile{
			rc: obj,
			pf: pf,
		}, nil
	} else {
		return nil, errors.New("fs not implement io.ReaderAt")
	}
}

func (p *ParquetGoReader) readRowsFromGroup(reader groupReader) ([]parquet.Row, error) {
	// ReadRows(buffer) method of the parquet reader reads up to len(buffer) rows into the buffer.
	// However, it does not guarantee that exactly len(buffer) rows will be read, so we must use the
	// returned number of rows (n) to check if we've reached the limit.
	rows := make([]parquet.Row, 0, reader.limit)
	tmp := make([]parquet.Row, 300)
	count := 0
	for {
		n, err := reader.reader.ReadRows(tmp)
		if n > 0 {
			rows = append(rows, tmp[:n]...)
			count += n
		}
		// break if we reach limit, or reach end of file
		if count > int(reader.limit) {
			rows = rows[:reader.limit]
			break
		}
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, err
		}
	}

	return rows, nil
}

func (p *ParquetGoReader) Rows(ctx context.Context, limit, offset int64) ([][]any, error) {
	if len(p.files) == 0 {
		return [][]any{}, nil
	}

	loader, err := p.loader(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	defer loader.close()

	data := [][]any{}
	for _, reader := range loader.groupReaders {
		rows, err := p.readRowsFromGroup(reader)
		if err != nil {
			return nil, err
		}

		for i, row := range rows {
			if i < int(reader.offset) {
				continue
			}
			rr := []any{}
			for _, cell := range row {
				rr = append(rr, cell.String())
			}
			data = append(data, rr)
		}
	}
	return data, nil
}

func (p *ParquetGoReader) Columns() []string {
	return p.columns
}

func (p *ParquetGoReader) Total(ctx context.Context) (int64, error) {
	files, err := p.loadFiles(ctx)
	if err != nil {
		return 0, err
	}
	defer func() {
		for _, f := range files {
			_ = f.rc.Close()
		}
	}()

	var total int64
	for _, f := range files {
		total += f.pf.NumRows()
	}

	return total, nil
}
