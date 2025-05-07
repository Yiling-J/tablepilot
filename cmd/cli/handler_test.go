package cli

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Yiling-J/tablepilot/config"
	"github.com/Yiling-J/tablepilot/ent"
	"github.com/Yiling-J/tablepilot/ent/schema"
	"github.com/Yiling-J/tablepilot/infra/db"
	"github.com/Yiling-J/tablepilot/services"
	"github.com/Yiling-J/tablepilot/services/table"
	"github.com/Yiling-J/tablepilot/utils/tableprinter"

	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestHandler_Create(t *testing.T) {
	tableMock := &table.TableServiceMock{
		CreateFunc: func(ctx context.Context, req *table.TableGenRequest) (string, error) {
			require.Equal(t, &table.TableGenRequest{
				Name: "go",
				Columns: []table.TableGenColumn{
					{Name: "c1", Type: "string", FillMode: "ai"},
				},
			}, req)
			return "123", nil
		},
	}
	handler := &Handler{
		backend: services.NewBackend(
			&config.Config{}, nil, zap.NewNop().Sugar(),
			nil, tableMock, nil,
		),
	}
	cmd := &cobra.Command{}
	testFile := fmt.Sprintf("foo_%d.json", time.Now().UnixNano())
	file, err := os.Create(testFile)
	require.NoError(t, err)
	defer os.Remove(testFile)
	_, err = file.WriteString(
		`{"name":"go","columns":[{"name":"c1","type":"string","fill_mode":"ai"}]}`,
	)
	require.NoError(t, err)
	err = file.Close()
	require.NoError(t, err)
	err = handler.Create(cmd, []string{testFile})
	require.NoError(t, err)
	require.NoError(t, err)
}

func TestHandler_Update(t *testing.T) {
	tableMock := &table.TableServiceMock{
		UpdateFunc: func(ctx context.Context, tb string, req *table.TableGenRequest) (string, error) {
			require.Equal(t, "foo", tb)
			require.Equal(t, &table.TableGenRequest{
				Name: "go",
				Columns: []table.TableGenColumn{
					{Name: "c1", Type: "string", FillMode: "ai"},
				},
			}, req)
			return "123", nil
		},
	}
	handler := &Handler{
		backend: services.NewBackend(
			&config.Config{}, nil, zap.NewNop().Sugar(),
			nil, tableMock, nil,
		),
	}
	cmd := &cobra.Command{}
	cmd.Flags().String("table", "", "")
	err := cmd.Flags().Set("table", "foo")
	require.NoError(t, err)
	testFile := fmt.Sprintf("foo_%d.json", time.Now().UnixNano())
	file, err := os.Create(testFile)
	require.NoError(t, err)
	defer os.Remove(testFile)
	_, err = file.WriteString(
		`{"name":"go","columns":[{"name":"c1","type":"string","fill_mode":"ai"}]}`,
	)
	require.NoError(t, err)
	err = file.Close()
	require.NoError(t, err)
	err = handler.Update(cmd, []string{testFile})
	require.NoError(t, err)
	require.NoError(t, err)
}

func TestHandler_Show(t *testing.T) {
	tableMock := &table.TableServiceMock{
		RowsFunc: func(ctx context.Context, name string) (*table.Rows, error) {
			require.Equal(t, "foo", name)
			return &table.Rows{
				Columns: []*ent.TableColumn{
					{Nanoid: "1", Name: "c1"},
					{Nanoid: "2", Name: "c2"},
				},
				Rows: []*ent.TableRow{
					{Cells: []*schema.CellValue{{Value: "a1"}, {Value: "b1"}}},
					{Cells: []*schema.CellValue{{Value: "a2"}, {Value: "b2"}}},
					{Cells: []*schema.CellValue{{Value: "a3"}, {Value: "b3"}}},
				},
			}, nil
		},
	}
	printer := &tableprinter.TablePrinterMock{
		AddHeaderFunc: func(strings []string, fieldOptionMoqParams ...tableprinter.FieldOption) {},
		AddFieldFunc:  func(s string, fieldOptions ...tableprinter.FieldOption) {},
		EndRowFunc:    func() {},
		RenderFunc:    func() error { return nil },
	}
	handler := NewHandler(
		services.NewBackend(
			&config.Config{}, nil, zap.NewNop().Sugar(),
			nil, tableMock, nil,
		),
	)
	handler.getPrinter = func() tableprinter.TablePrinter { return printer }
	cmd := &cobra.Command{}
	err := handler.Show(cmd, []string{"foo"})
	require.NoError(t, err)
	require.Equal(t, 1, len(printer.AddHeaderCalls()))
	require.Equal(t, []string{"c1", "c2"}, printer.AddHeaderCalls()[0].Strings)
	require.Equal(t, 6, len(printer.AddFieldCalls()))
	fields := []string{}
	for _, call := range printer.AddFieldCalls() {
		fields = append(fields, call.S)
	}
	require.Equal(t, []string{"a1", "b1", "a2", "b2", "a3", "b3"}, fields)
	require.Equal(t, 3, len(printer.EndRowCalls()))
	require.Equal(t, 1, len(printer.RenderCalls()))
}

func TestHandler_List(t *testing.T) {
	tableMock := &table.TableServiceMock{
		ListTablesFunc: func(ctx context.Context) (*table.ListTablesResponse, error) {
			return &table.ListTablesResponse{
				Total: 2,
				Tables: []table.TableInfo{
					{ID: "1", Name: "t1", Description: "d1"},
					{ID: "2", Name: "t2", Description: "d2"},
				},
			}, nil
		},
	}
	printer := &tableprinter.TablePrinterMock{
		AddHeaderFunc: func(strings []string, fieldOptionMoqParams ...tableprinter.FieldOption) {},
		AddFieldFunc:  func(s string, fieldOptions ...tableprinter.FieldOption) {},
		EndRowFunc:    func() {},
		RenderFunc:    func() error { return nil },
	}
	handler := NewHandler(
		services.NewBackend(
			&config.Config{}, nil, zap.NewNop().Sugar(),
			nil, tableMock, nil,
		),
	)
	handler.getPrinter = func() tableprinter.TablePrinter { return printer }
	cmd := &cobra.Command{}
	err := handler.List(cmd, []string{})
	require.NoError(t, err)
	require.Equal(t, 1, len(printer.AddHeaderCalls()))
	require.Equal(t, []string{"ID", "Name", "Description"}, printer.AddHeaderCalls()[0].Strings)
	require.Equal(t, 6, len(printer.AddFieldCalls()))
	fields := []string{}
	for _, call := range printer.AddFieldCalls() {
		fields = append(fields, call.S)
	}
	require.Equal(t, []string{"1", "t1", "d1", "2", "t2", "d2"}, fields)
	require.Equal(t, 2, len(printer.EndRowCalls()))
	require.Equal(t, 1, len(printer.RenderCalls()))
}

func TestHandler_Delete(t *testing.T) {
	tableMock := &table.TableServiceMock{
		DeleteFunc: func(ctx context.Context, table string) (int, error) {
			require.Equal(t, "foo", table)
			return 1, nil
		},
	}
	handler := NewHandler(
		services.NewBackend(
			&config.Config{}, nil, zap.NewNop().Sugar(),
			nil, tableMock, nil,
		),
	)
	cmd := &cobra.Command{}
	err := handler.Delete(cmd, []string{"foo"})
	require.NoError(t, err)
}

func TestHandler_Export(t *testing.T) {
	for _, toFlag := range []bool{false, true} {
		t.Run(fmt.Sprintf("to flag set %v", toFlag), func(t *testing.T) {
			tableMock := &table.TableServiceMock{
				RowsFunc: func(ctx context.Context, name string) (*table.Rows, error) {
					require.Equal(t, "foo", name)
					return &table.Rows{
						Columns: []*ent.TableColumn{
							{Nanoid: "1", Name: "c1"},
							{Nanoid: "2", Name: "c2"},
						},
						Rows: []*ent.TableRow{
							{Cells: []*schema.CellValue{
								{Value: "a1"}, {Value: "b1"},
							}},
							{Cells: []*schema.CellValue{
								{Value: "a2"}, {Value: "b2"},
							}},
							{Cells: []*schema.CellValue{
								{Value: "a3"}, {Value: "b3"},
							}},
						},
					}, nil
				},
			}
			handler := NewHandler(
				services.NewBackend(
					&config.Config{}, nil, zap.NewNop().Sugar(),
					nil, tableMock, nil,
				),
			)
			cmd := &cobra.Command{}
			cmd.Flags().StringP("to", "", "", "")
			if toFlag {
				err := cmd.Flags().Set("to", "foo_export.csv")
				require.NoError(t, err)
			}
			err := handler.Export(cmd, []string{"foo"})
			require.NoError(t, err)
			files, err := os.ReadDir(".")
			require.NoError(t, err)
			csvFile := "foo_export.csv"
			if !toFlag {
				for _, f := range files {
					if strings.HasPrefix(f.Name(), "foo_") && strings.HasSuffix(f.Name(), ".csv") {
						csvFile = f.Name()
						break
					}
				}
			}
			require.NotEmpty(t, csvFile)
			defer os.Remove(csvFile)

			file, err := os.Open(csvFile)
			require.NoError(t, err)
			defer file.Close()
			reader := csv.NewReader(file)
			records, err := reader.ReadAll()
			require.NoError(t, err)
			require.Equal(
				t,
				[][]string{
					{"c1", "c2"},
					{"a1", "b1"},
					{"a2", "b2"},
					{"a3", "b3"},
				}, records)
		})
	}
}

func TestHandler_Generate(t *testing.T) {
	for _, saveTo := range []bool{false, true} {
		t.Run(fmt.Sprintf("saveto flag set %v", saveTo), func(t *testing.T) {
			var counter int
			mockRowGen := &table.RowsGeneratorMock{
				NextFunc: func(ctx context.Context) ([]map[string]*schema.CellValue, error) {
					defer func() { counter += 1 }()
					if counter < 2 {
						return []map[string]*schema.CellValue{
							{
								"1": &schema.CellValue{Value: cast.ToString(counter)},
								"2": &schema.CellValue{Value: "t" + cast.ToString(counter)},
							},
						}, nil
					}
					return []map[string]*schema.CellValue{}, nil
				},
				TableFunc: func() *ent.TableMeta {
					return &ent.TableMeta{
						Name: "foo",
						Edges: ent.TableMetaEdges{
							Columns: []*ent.TableColumn{
								{Nanoid: "1", Name: "c1"},
								{Nanoid: "2", Name: "c2"},
							},
						},
					}
				},
			}
			tableMock := &table.TableServiceMock{
				GenetateFunc: func(ctx context.Context, params table.GenerateRowsRequest) (table.RowsGenerator, error) {
					require.Equal(t, "foo", params.Table)
					if saveTo {
						require.Equal(t, "foo_gen.csv", params.SaveTo)
					} else {
						require.Equal(t, "", params.SaveTo)
					}
					require.Equal(t, 4, params.Count)
					require.Equal(t, 2, params.Batch)
					require.Equal(t, 0.56, params.Temperature)
					require.Equal(t, "aiai", params.Model)
					return mockRowGen, nil
				},
			}
			printer := &tableprinter.TablePrinterMock{
				AddHeaderFunc: func(strings []string, fieldOptionMoqParams ...tableprinter.FieldOption) {},
				AddFieldFunc:  func(s string, fieldOptions ...tableprinter.FieldOption) {},
				EndRowFunc:    func() {},
				RenderFunc:    func() error { return nil },
			}
			handler := NewHandler(
				services.NewBackend(
					&config.Config{}, nil, zap.NewNop().Sugar(),
					nil, tableMock, nil,
				),
			)
			handler.getPrinter = func() tableprinter.TablePrinter { return printer }
			cmd := &cobra.Command{}
			cmd.Flags().IntP("count", "", 0, "")
			cmd.Flags().IntP("batch", "", 0, "")
			cmd.Flags().StringP("saveto", "s", "", "")
			cmd.Flags().Float64P("temperature", "", 0.6, "")
			cmd.Flags().StringP("model", "", "", "")
			err := cmd.Flags().Set("count", "4")
			require.NoError(t, err)
			err = cmd.Flags().Set("batch", "2")
			require.NoError(t, err)
			if saveTo {
				err = cmd.Flags().Set("saveto", "foo_gen.csv")
				require.NoError(t, err)
			}
			err = cmd.Flags().Set("temperature", "0.56")
			require.NoError(t, err)

			err = cmd.Flags().Set("model", "aiai")
			require.NoError(t, err)

			err = handler.Generate(cmd, []string{"foo"})
			require.NoError(t, err)
			if saveTo {
				defer os.Remove("foo_gen.csv")
			}

			require.Equal(t, 1, len(printer.AddHeaderCalls()))
			require.Equal(t, []string{"c1", "c2"}, printer.AddHeaderCalls()[0].Strings)
			require.Equal(t, 4, len(printer.AddFieldCalls()))
			fields := []string{}
			for _, call := range printer.AddFieldCalls() {
				fields = append(fields, call.S)
			}
			require.Equal(t, []string{"0", "t0", "1", "t1"}, fields)
			require.Equal(t, 2, len(printer.EndRowCalls()))
			require.Equal(t, 2, len(printer.RenderCalls()))

			if saveTo {
				file, err := os.Open("foo_gen.csv")
				require.NoError(t, err)
				defer file.Close()
				reader := csv.NewReader(file)
				records, err := reader.ReadAll()
				require.NoError(t, err)
				require.Equal(
					t,
					[][]string{{"c1", "c2"}, {"0", "t0"}, {"1", "t1"}},
					records)
			}
		})
	}
}

func TestHandler_Import(t *testing.T) {
	for _, name := range []string{"", "bar"} {
		t.Run(name, func(t *testing.T) {
			file, err := os.Create("foo.csv")
			require.NoError(t, err)
			defer file.Close()
			defer os.Remove("foo.csv")
			writer := csv.NewWriter(file)
			data := [][]string{
				{"c1", "c2"},
				{"v1", "v2"},
			}
			err = writer.WriteAll(data)
			require.NoError(t, err)
			writer.Flush()

			tableMock := &table.TableServiceMock{
				ImportFunc: func(ctx context.Context, table string, reader io.Reader) (string, error) {
					if name == "" {
						require.Equal(t, "foo", table)
					} else {
						require.Equal(t, name, table)
					}
					b, err := io.ReadAll(reader)
					require.NoError(t, err)
					require.Equal(t, "c1,c2\nv1,v2\n", string(b))
					return "123", nil
				},
			}
			handler := NewHandler(
				services.NewBackend(
					&config.Config{}, nil, zap.NewNop().Sugar(),
					nil, tableMock, nil,
				),
			)
			cmd := &cobra.Command{}
			cmd.Flags().String("table", "", "")
			if name != "" {
				err = cmd.Flags().Set("table", name)
				require.NoError(t, err)
			}
			err = handler.Import(cmd, []string{"foo.csv"})
			require.NoError(t, err)
		})
	}
}

func TestHandler_Truncate(t *testing.T) {
	tableMock := &table.TableServiceMock{
		TruncateFunc: func(ctx context.Context, table string) (int, error) {
			require.Equal(t, "foo", table)
			return 5, nil
		},
	}
	handler := NewHandler(
		services.NewBackend(
			&config.Config{}, nil, zap.NewNop().Sugar(),
			nil, tableMock, nil,
		),
	)
	cmd := &cobra.Command{}
	err := handler.Truncate(cmd, []string{"foo"})
	require.NoError(t, err)
}

func TestHandler_Describe(t *testing.T) {
	tableMock := &table.TableServiceMock{
		GetTableDetailFunc: func(ctx context.Context, name string) (*table.TableInfo, error) {
			require.Equal(t, "foo", name)
			return &table.TableInfo{
				ID: "t1",
				Columns: []table.TableColumnInfo{
					{ID: "1", Name: "c1", Type: "string", FillMode: "ai", Description: "d1"},
					{ID: "2", Name: "c2", Type: "int", FillMode: "bi", Description: "d2"},
				}}, nil
		},
		GetTableSchemaFunc: func(ctx context.Context, tb string) (*table.TableGenRequest, error) {
			require.Equal(t, "foo", tb)
			return &table.TableGenRequest{
				Name: "abc",
			}, nil
		},
	}
	printer := &tableprinter.TablePrinterMock{
		AddHeaderFunc: func(strings []string, fieldOptionMoqParams ...tableprinter.FieldOption) {},
		AddFieldFunc:  func(s string, fieldOptions ...tableprinter.FieldOption) {},
		EndRowFunc:    func() {},
		RenderFunc:    func() error { return nil },
	}
	handler := NewHandler(
		services.NewBackend(
			&config.Config{}, nil, zap.NewNop().Sugar(),
			nil, tableMock, nil,
		),
	)
	handler.getPrinter = func() tableprinter.TablePrinter { return printer }
	cmd := &cobra.Command{}
	cmd.Flags().String("output", "table", "")
	err := handler.Describe(cmd, []string{"foo"})
	require.NoError(t, err)
	require.Equal(t, 1, len(printer.AddHeaderCalls()))
	require.Equal(t, []string{"ID", "Name", "Type", "Fill Mode", "Description"}, printer.AddHeaderCalls()[0].Strings)
	require.Equal(t, 10, len(printer.AddFieldCalls()))
	fields := []string{}
	for _, call := range printer.AddFieldCalls() {
		fields = append(fields, call.S)
	}
	require.Equal(t, []string{"1", "c1", "string", "ai", "d1", "2", "c2", "int", "bi", "d2"}, fields)
	require.Equal(t, 2, len(printer.EndRowCalls()))
	require.Equal(t, 1, len(printer.RenderCalls()))

	err = cmd.Flags().Set("output", "json")
	require.NoError(t, err)
	err = handler.Describe(cmd, []string{"foo"})
	require.NoError(t, err)
	require.Equal(t, 1, len(tableMock.GetTableSchemaCalls()))
}

func TestHandler_Autofill(t *testing.T) {
	for _, saveTo := range []bool{false, true} {
		t.Run(fmt.Sprintf("saveto flag set %v", saveTo), func(t *testing.T) {
			var counter int
			mockRowGen := &table.RowsGeneratorMock{
				NextFunc: func(ctx context.Context) ([]map[string]*schema.CellValue, error) {
					defer func() { counter += 1 }()
					if counter < 2 {
						return []map[string]*schema.CellValue{
							{
								"1": &schema.CellValue{Value: cast.ToString(counter)},
								"2": &schema.CellValue{Value: "t" + cast.ToString(counter)},
							},
						}, nil
					}
					return []map[string]*schema.CellValue{}, nil
				},
				TableFunc: func() *ent.TableMeta {
					return &ent.TableMeta{
						Name: "foo",
						Edges: ent.TableMetaEdges{
							Columns: []*ent.TableColumn{
								{Nanoid: "1", Name: "c1"},
								{Nanoid: "2", Name: "c2"},
							},
						},
					}
				},
			}
			tableMock := &table.TableServiceMock{
				GenetateFunc: func(ctx context.Context, params table.GenerateRowsRequest) (table.RowsGenerator, error) {
					require.Equal(t, "foo", params.Table)
					if saveTo {
						require.Equal(t, "foo_gen.csv", params.SaveTo)
					} else {
						require.Equal(t, "", params.SaveTo)
					}
					require.Equal(t, 4, params.Count)
					require.Equal(t, 2, params.Batch)
					require.Equal(t, 0.56, params.Temperature)
					require.Equal(t, "aiai", params.Model)
					require.Equal(t, true, params.Autofill.Enable)
					require.Equal(t, 3, params.Autofill.Offset)
					require.Equal(t, []string{"c1"}, params.Autofill.Columns)
					require.Equal(t, []string{"c2"}, params.Autofill.ContextColumns)
					return mockRowGen, nil
				},
			}
			printer := &tableprinter.TablePrinterMock{
				AddHeaderFunc: func(strings []string, fieldOptionMoqParams ...tableprinter.FieldOption) {},
				AddFieldFunc:  func(s string, fieldOptions ...tableprinter.FieldOption) {},
				EndRowFunc:    func() {},
				RenderFunc:    func() error { return nil },
			}
			handler := NewHandler(
				services.NewBackend(
					&config.Config{}, nil, zap.NewNop().Sugar(),
					nil, tableMock, nil,
				),
			)
			handler.getPrinter = func() tableprinter.TablePrinter { return printer }
			cmd := &cobra.Command{}
			cmd.Flags().IntP("count", "", 0, "")
			cmd.Flags().IntP("batch", "", 0, "")
			cmd.Flags().StringP("saveto", "s", "", "")
			cmd.Flags().Float64P("temperature", "", 0.6, "")
			cmd.Flags().StringP("model", "", "", "")
			cmd.Flags().Int("offset", 3, "")
			cmd.Flags().StringArray("columns", []string{}, "")
			cmd.Flags().StringArray("context_columns", []string{}, "")
			err := cmd.Flags().Set("count", "4")
			require.NoError(t, err)
			err = cmd.Flags().Set("batch", "2")
			require.NoError(t, err)
			if saveTo {
				err = cmd.Flags().Set("saveto", "foo_gen.csv")
				require.NoError(t, err)
			}
			err = cmd.Flags().Set("temperature", "0.56")
			require.NoError(t, err)

			err = cmd.Flags().Set("model", "aiai")
			require.NoError(t, err)

			err = cmd.Flags().Set("columns", "c1")
			require.NoError(t, err)

			err = cmd.Flags().Set("context_columns", "c2")
			require.NoError(t, err)

			err = handler.Autofill(cmd, []string{"foo"})
			require.NoError(t, err)
			if saveTo {
				defer os.Remove("foo_gen.csv")
			}

			require.Equal(t, 1, len(printer.AddHeaderCalls()))
			require.Equal(t, []string{"c1", "c2"}, printer.AddHeaderCalls()[0].Strings)
			require.Equal(t, 4, len(printer.AddFieldCalls()))
			fields := []string{}
			for _, call := range printer.AddFieldCalls() {
				fields = append(fields, call.S)
			}
			require.Equal(t, []string{"0", "t0", "1", "t1"}, fields)
			require.Equal(t, 2, len(printer.EndRowCalls()))
			require.Equal(t, 2, len(printer.RenderCalls()))

			if saveTo {
				file, err := os.Open("foo_gen.csv")
				require.NoError(t, err)
				defer file.Close()
				reader := csv.NewReader(file)
				records, err := reader.ReadAll()
				require.NoError(t, err)
				require.Equal(
					t,
					[][]string{{"c1", "c2"}, {"0", "t0"}, {"1", "t1"}},
					records)
			}
		})
	}
}

func TestHandler_Builder(t *testing.T) {
	tableMock := &table.TableServiceMock{
		GenerateBuilderTablesFunc: func(ctx context.Context, prompt string, params table.ModelParams) ([]table.BuilderTable, error) {
			return []table.BuilderTable{
				{Name: "tb1", Description: "d1"},
			}, nil
		},
		PolishBuilderTablesFunc: func(ctx context.Context, tables []table.BuilderTable, prompt string, params table.ModelParams) ([]table.BuilderTable, error) {
			return []table.BuilderTable{
				{Name: "tb1", Description: "d1"},
				{Name: "tb2", Description: "d2", Depends: []string{"tb1"}},
			}, nil
		},
		BuildTableFunc: func(ctx context.Context, name, description string, depends []string, exists []*table.TableInfo, params table.ModelParams) (*table.TableGenRequest, error) {
			switch name {
			case "tb1":
				require.Equal(t, 0, len(exists))
				return &table.TableGenRequest{
					Name:        "tb1",
					Description: "d1",
					Columns:     []table.TableGenColumn{},
				}, nil
			case "tb2":
				require.Equal(t, []*table.TableInfo{
					{Name: "tbb1"},
				}, exists)
				return &table.TableGenRequest{
					Name:        "tb2",
					Description: "d2",
					Columns:     []table.TableGenColumn{},
				}, nil
			}
			return nil, errors.New("build table err")
		},
		PolishBuilderTableFunc: func(ctx context.Context, req *table.TableGenRequest, prompt string, exists []*table.TableInfo, params table.ModelParams) (*table.TableGenRequest, error) {
			return &table.TableGenRequest{
				Name:        "tb1",
				Description: "d1",
				Columns: []table.TableGenColumn{
					{Name: "col1", Description: "dc1"},
				},
			}, nil
		},
		CreateFunc: func(ctx context.Context, req *table.TableGenRequest) (string, error) {
			switch req.Name {
			case "tb1":
				return "id1", nil
			case "tb2":
				return "id2", nil
			default:
				return "", errors.New("create err")
			}
		},
		GetTableDetailFunc: func(ctx context.Context, tb string) (*table.TableInfo, error) {
			name := ""
			switch tb {
			case "id1":
				name = "tbb1"
			case "id2":
				name = "tbb2"
			default:
				return nil, errors.New("get table detail err")
			}
			return &table.TableInfo{Name: name}, nil
		},
		ValidateFunc: func(ctx context.Context, req *table.TableGenRequest) error {
			return nil
		},
	}
	handler := NewHandler(
		services.NewBackend(
			&config.Config{}, db.NewTestDB(), zap.NewNop().Sugar(),
			nil, tableMock, nil,
		),
	)
	cmd := &cobra.Command{}
	cmd.Flags().Float64P("temperature", "", 0.3, "")
	cmd.Flags().StringP("model", "", "", "")
	cmd.SetContext(context.TODO())
	cmd.SetIn(bytes.NewReader([]byte("foo\nbar\n\nbaz\n\n\n")))
	err := handler.Builder(cmd, []string{})
	require.NoError(t, err)
	require.Equal(t, 1, len(tableMock.GenerateBuilderTablesCalls()))
	require.Equal(t, 1, len(tableMock.PolishBuilderTablesCalls()))
	require.Equal(t, 2, len(tableMock.BuildTableCalls()))
	require.Equal(t, 1, len(tableMock.PolishBuilderTableCalls()))
	require.Equal(t, 2, len(tableMock.CreateCalls()))
	require.Equal(t, 2, len(tableMock.GetTableDetailCalls()))
}

func TestHandler_TopoSortTables(t *testing.T) {
	tests := []struct {
		name     string
		tables   []table.BuilderTable
		wantErr  bool
		validate func(sorted []table.BuilderTable, err error)
	}{
		{
			name: "No dependency",
			tables: []table.BuilderTable{
				{Name: "A", Depends: []string{}},
				{Name: "B", Depends: []string{}},
				{Name: "C", Depends: []string{}},
			},
			wantErr: false,
			validate: func(sorted []table.BuilderTable, err error) {
				require.NoError(t, err)
				nameIndex := map[string]bool{}
				for _, tbl := range sorted {
					nameIndex[tbl.Name] = true
				}
				require.True(t, nameIndex["A"])
				require.True(t, nameIndex["B"])
				require.True(t, nameIndex["C"])
			},
		},
		{
			name: "Basic linear dependency",
			tables: []table.BuilderTable{
				{Name: "A", Depends: []string{}},
				{Name: "B", Depends: []string{"A"}},
				{Name: "C", Depends: []string{"B"}},
			},
			wantErr: false,
			validate: func(sorted []table.BuilderTable, err error) {
				require.NoError(t, err)
				nameIndex := map[string]int{}
				for i, tbl := range sorted {
					nameIndex[tbl.Name] = i
				}
				require.True(t, nameIndex["A"] < nameIndex["B"])
				require.True(t, nameIndex["B"] < nameIndex["C"])
			},
		},
		{
			name: "Branching dependencies",
			tables: []table.BuilderTable{
				{Name: "A", Depends: []string{}},
				{Name: "B", Depends: []string{"A"}},
				{Name: "C", Depends: []string{"A"}},
				{Name: "D", Depends: []string{"B", "C"}},
			},
			wantErr: false,
			validate: func(sorted []table.BuilderTable, err error) {
				require.NoError(t, err)
				nameIndex := map[string]int{}
				for i, tbl := range sorted {
					nameIndex[tbl.Name] = i
				}
				require.True(t, nameIndex["A"] < nameIndex["B"])
				require.True(t, nameIndex["A"] < nameIndex["C"])
				require.True(t, nameIndex["B"] < nameIndex["D"])
				require.True(t, nameIndex["C"] < nameIndex["D"])
			},
		},
		{
			name: "Cyclic dependencies",
			tables: []table.BuilderTable{
				{Name: "A", Depends: []string{"C"}},
				{Name: "B", Depends: []string{"A"}},
				{Name: "C", Depends: []string{"B"}}, // cycle: A -> C -> B -> A
			},
			wantErr: true,
			validate: func(sorted []table.BuilderTable, err error) {
				require.Error(t, err)
			},
		},
		{
			name:    "Empty input",
			tables:  []table.BuilderTable{},
			wantErr: false,
			validate: func(sorted []table.BuilderTable, err error) {
				require.NoError(t, err)
				require.Len(t, sorted, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the function under test
			sorted, err := topoSortTables(tt.tables)

			// Check if we expect an error
			if tt.wantErr {
				require.Error(t, err)
				tt.validate(sorted, err)
			} else {
				require.NoError(t, err)
				tt.validate(sorted, err)
			}
		})
	}
}
