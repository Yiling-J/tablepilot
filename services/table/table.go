package table

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/Yiling-J/tablepilot/config"
	"github.com/Yiling-J/tablepilot/ent"
	"github.com/Yiling-J/tablepilot/ent/schema"
	"github.com/Yiling-J/tablepilot/ent/tablecolumn"
	"github.com/Yiling-J/tablepilot/ent/tablemeta"
	"github.com/Yiling-J/tablepilot/ent/tablerow"
	"github.com/Yiling-J/tablepilot/services/ai"
	"github.com/Yiling-J/tablepilot/services/ai/client"
	"github.com/Yiling-J/tablepilot/services/ai/promptbuilder"
	"github.com/Yiling-J/tablepilot/services/table/source"
	"github.com/Yiling-J/tablepilot/services/table/util"

	"github.com/invopop/jsonschema"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
)

const (
	CREATE_TABLE_TEMPERATURE  = 0.1
	CREATE_TABLE_MAX_TOKENS   = 3000
	GENERATE_DATA_TEMPERATURE = 0.6
	GENERATE_DATA_MAX_TOKENS  = 6000
)

var reflector = jsonschema.Reflector{
	AllowAdditionalProperties: false,
	DoNotReference:            true,
}

//go:generate moq -rm -out table_moq.go . TableService RowsGenerator
type TableService interface {
	CreateTable(ctx context.Context, req *TableGenRequest) (string, error)
	ListTables(ctx context.Context) (*ListTablesResponse, error)
	GetTableDetail(ctx context.Context, table string) (*TableInfo, error)
	Genetate(ctx context.Context, params GenerateRowsRequest) (RowsGenerator, error)
	Rows(ctx context.Context, table string) (*Rows, error)
	Truncate(ctx context.Context, table string) (int, error)
	Delete(ctx context.Context, table string) (int, error)
	Import(ctx context.Context, table string, reader io.Reader) (string, error)
	CreateRows(ctx context.Context, table string, rows []map[string]any) error
}

type TableServiceImpl struct {
	db     *ent.Client
	ai     ai.AiService
	logger *zap.SugaredLogger
}

func NewTableService(config *config.Config, db *ent.Client, ai ai.AiService, logger *zap.SugaredLogger) *TableServiceImpl {
	ts := &TableServiceImpl{
		db:     db,
		ai:     ai,
		logger: logger.With("service", "table"),
	}
	return ts
}

type TableGenColumnSchema struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
}

func (t *TableServiceImpl) CreateTable(ctx context.Context, req *TableGenRequest) (string, error) {
	t.logger.Debugw("creating table", "name", req.Name)
	tx, err := t.db.Tx(ctx)
	if err != nil {
		return "", fmt.Errorf("starting a transaction: %w", err)
	}
	if req.Name == "" {
		return "", ent.Rollback(tx, errors.New("table name is empty"))
	}
	sources := map[string]source.Source{}
	if len(req.Sources) > 0 {
		for _, raw := range req.Sources {
			s, err := source.ValidateSource(ctx, raw, tx.Client())
			if err != nil {
				return "", ent.Rollback(tx, err)
			}
			sources[gjson.GetBytes(raw, "name").String()] = s
		}
	}

	table, err := tx.TableMeta.Create().SetName(req.Name).SetDescription(req.Description).SetModel(req.Model).Save(ctx)
	if err != nil {
		return "", ent.Rollback(tx, err)
	}
	extra := 0
	columns := []TableGenColumn{}
	for _, col := range req.Columns {
		if col.Name == "" {
			extra += 1
		} else {
			columns = append(columns, col)
		}
	}
	if extra > 0 {
		t.logger.Info("start generating extra columns")
		columnsGenBuilder := promptbuilder.NewColumnsBuilder(extra, req.Name, req.Description)
		bc := []promptbuilder.Column{}
		for _, c := range columns {
			bc = append(bc, promptbuilder.Column{
				Name:        c.Name,
				Description: c.Description,
			})
		}
		columnsGenBuilder.AddExistingColumns(bc)
		prompt, err := columnsGenBuilder.Prompt()
		if err != nil {
			return "", err
		}
		resp, err := t.ai.Chat(ctx, &client.ChatRequest{
			Messages: []*client.Message{
				client.UserMessage(prompt),
			},
			Model:           table.Model,
			Temperature:     CREATE_TABLE_TEMPERATURE,
			MaxOutputTokens: CREATE_TABLE_MAX_TOKENS,
			Schema:          reflector.Reflect([]TableGenColumnSchema{}),
		})
		if err != nil {
			return "", err
		}
		t.logger.Debug("extra columns generated")
		extraColumns, err := util.TryDecodeJsonArray[TableGenColumn](resp.Content)
		if err != nil && len(extraColumns) == 0 {
			return "", err
		}
		if len(extraColumns) > extra {
			extraColumns = extraColumns[:extra]
		}
		if err != nil {
			t.logger.Errorw("err happen when decode column gen response json", "err", err)
		}
		for i := range extraColumns {
			extraColumns[i].FillMode = "ai"
		}
		columns = append(columns, extraColumns...)
	}
	columnCreates := []*ent.TableColumnCreate{}
	for _, col := range columns {
		cc := tx.TableColumn.Create().
			SetTablemeta(table).
			SetFillMode(tablecolumn.FillMode(col.FillMode)).
			SetName(col.Name).
			SetDescription(col.Description).
			SetType(tablecolumn.Type(col.Type)).
			SetContextLength(col.ContextLength)

		if col.Source != "" && col.FillMode == "pick" {
			if s, ok := sources[col.Source]; ok {
				v, err := json.Marshal(s)
				if err != nil {
					return "", ent.Rollback(tx, err)
				}
				cc.SetSource(v)
			} else {
				return "", ent.Rollback(tx, fmt.Errorf("source %s not dound", col.Source))
			}
		}
		columnCreates = append(columnCreates, cc)
		t.logger.Debugw(
			"creating column", "name", col.Name, "description", col.Description,
			"fill_mode", col.FillMode, "type", col.Type, "source", col.Source,
		)
	}

	err = tx.TableColumn.CreateBulk(columnCreates...).Exec(ctx)
	if err != nil {
		return "", ent.Rollback(tx, err)
	}

	t.logger.Debug("finish creating table")
	return table.Nanoid, tx.Commit()
}

func (t *TableServiceImpl) ListTables(ctx context.Context) (*ListTablesResponse, error) {
	tables, err := t.db.TableMeta.Query().WithColumns(func(tcq *ent.TableColumnQuery) {
		tcq.Order(ent.Asc(tablecolumn.FieldID))
	}).Order(ent.Desc(tablemeta.FieldID)).All(ctx)
	if err != nil {
		return nil, err
	}
	resp := &ListTablesResponse{Total: len(tables)}
	for _, table := range tables {
		columns := []TableColumnInfo{}
		for _, column := range table.Edges.Columns {
			columns = append(columns, TableColumnInfo{
				ID:          column.Nanoid,
				Name:        column.Name,
				Description: column.Description,
				Type:        column.Type.String(),
				FillMode:    column.FillMode.String(),
			})
		}
		resp.Tables = append(resp.Tables, TableInfo{
			ID:          table.Nanoid,
			Name:        table.Name,
			Description: table.Description,
			Model:       table.Model,
			Columns:     columns,
		})
	}
	return resp, nil
}

func (t *TableServiceImpl) GetTableDetail(ctx context.Context, table string) (*TableInfo, error) {
	meta, err := t.db.TableMeta.Query().WithColumns(func(tcq *ent.TableColumnQuery) {
		tcq.Order(ent.Asc(tablecolumn.FieldID))
	}).Where(tablemeta.Or(
		tablemeta.Nanoid(table),
		tablemeta.Name(table),
	)).First(ctx)
	if err != nil {
		return nil, err
	}
	columns := []TableColumnInfo{}
	for _, column := range meta.Edges.Columns {
		columns = append(columns, TableColumnInfo{
			ID:          column.Nanoid,
			Name:        column.Name,
			Description: column.Description,
			Type:        column.Type.String(),
			FillMode:    column.FillMode.String(),
		})
	}
	return &TableInfo{
		ID:          meta.Nanoid,
		Name:        meta.Name,
		Description: meta.Description,
		Model:       meta.Model,
		Columns:     columns,
	}, nil
}

func (t *TableServiceImpl) Genetate(ctx context.Context, params GenerateRowsRequest) (RowsGenerator, error) {
	return NewRowsGenerator(ctx, params, t.db, t.ai, t.logger)
}

func (t *TableServiceImpl) Rows(ctx context.Context, table string) (*Rows, error) {
	meta, err := t.db.TableMeta.Query().WithColumns(func(tcq *ent.TableColumnQuery) {
		tcq.Order(ent.Asc(tablecolumn.FieldID))
	}).WithRows(func(trq *ent.TableRowQuery) {
		trq.Order(ent.Asc(tablerow.FieldID))
	}).Where(tablemeta.Or(
		tablemeta.Nanoid(table),
		tablemeta.Name(table),
	)).First(ctx)
	if err != nil {
		return nil, err
	}
	return &Rows{Columns: meta.Edges.Columns, Rows: meta.Edges.Rows}, nil
}

func (t *TableServiceImpl) Truncate(ctx context.Context, table string) (int, error) {
	meta, err := t.db.TableMeta.Query().Where(tablemeta.Or(
		tablemeta.Nanoid(table),
		tablemeta.Name(table),
	)).First(ctx)
	if err != nil {
		return 0, err
	}
	return t.db.TableRow.Delete().Where(tablerow.HasTablemetaWith(tablemeta.Nanoid(meta.Nanoid))).Exec(ctx)
}

func (t *TableServiceImpl) Delete(ctx context.Context, table string) (int, error) {
	return t.db.TableMeta.Delete().Where(tablemeta.Or(
		tablemeta.Nanoid(table),
		tablemeta.Name(table),
	)).Exec(ctx)
}

func (t *TableServiceImpl) Import(ctx context.Context, table string, reader io.Reader) (string, error) {
	tx, err := t.db.Tx(ctx)
	if err != nil {
		return "", fmt.Errorf("starting a transaction: %w", err)
	}
	exists, err := tx.TableMeta.Query().WithColumns(func(tcq *ent.TableColumnQuery) {
		tcq.Order(ent.Asc(tablecolumn.FieldID))
	}).Where(
		tablemeta.Name(table),
	).All(ctx)
	if err != nil {
		return "", ent.Rollback(tx, err)
	}
	// if table already exists, import data to matched columns for the table
	var tablemeta *ent.TableMeta
	if len(exists) > 0 {
		tablemeta = exists[0]
	}
	cr := csv.NewReader(reader)
	columns := []string{}
	// rows read from csv
	rows := [][]string{}
	// rows import to table
	importRows := [][]any{}
	counter := 0
	for {
		record, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", ent.Rollback(tx, err)
		}
		if counter == 0 {
			columns = record
		} else {
			rows = append(rows, record)
		}
		counter += 1
	}

	if tablemeta != nil {
		cm := map[string]int{}
		for i, col := range columns {
			cm[col] = i
		}
		for _, row := range rows {
			newRow := []any{}
			for _, col := range tablemeta.Edges.Columns {
				if j, ok := cm[col.Name]; ok {
					v, err := util.ConvertStringToType(row[j], col.Type)
					if err != nil {
						return "", ent.Rollback(tx, err)
					}
					newRow = append(newRow, v)
				} else {
					v, err := util.ConvertStringToType("", col.Type)
					if err != nil {
						return "", ent.Rollback(tx, err)
					}
					newRow = append(newRow, v)
				}
			}
			importRows = append(importRows, newRow)
		}
	} else {
		for _, row := range rows {
			newRow := []any{}
			for _, v := range row {
				newRow = append(newRow, v)
			}
			importRows = append(importRows, newRow)
		}
		tablemeta, err = tx.TableMeta.Create().SetName(table).Save(ctx)
		if err != nil {
			return "", ent.Rollback(tx, err)
		}
		columnCreates := []*ent.TableColumnCreate{}
		for _, col := range columns {
			columnCreates = append(
				columnCreates, tx.TableColumn.Create().SetName(col).SetType(tablecolumn.TypeString).
					SetFillMode(tablecolumn.FillModeAi).SetTablemeta(tablemeta),
			)
		}
		err = tx.TableColumn.CreateBulk(columnCreates...).Exec(ctx)
		if err != nil {
			return "", ent.Rollback(tx, err)
		}
	}
	rowCreates := []*ent.TableRowCreate{}
	for _, row := range importRows {
		cells := []*schema.CellValue{}
		for _, cell := range row {
			cells = append(cells, &schema.CellValue{Value: cell})
		}
		rowCreates = append(rowCreates, tx.TableRow.Create().SetCells(cells).SetTablemeta(tablemeta))
	}
	err = tx.TableRow.CreateBulk(rowCreates...).Exec(ctx)
	if err != nil {
		return "", ent.Rollback(tx, err)
	}
	return tablemeta.Nanoid, tx.Commit()
}

func (t *TableServiceImpl) CreateRows(ctx context.Context, table string, rows []map[string]any) error {
	tx, err := t.db.Tx(ctx)
	if err != nil {
		return fmt.Errorf("starting a transaction: %w", err)
	}
	tablemeta, err := tx.TableMeta.Query().WithColumns(func(tcq *ent.TableColumnQuery) {
		tcq.Order(ent.Asc(tablecolumn.FieldID))
	}).Where(
		tablemeta.Or(
			tablemeta.Nanoid(table),
			tablemeta.Name(table)),
	).First(ctx)
	if err != nil {
		return ent.Rollback(tx, err)
	}
	importRows := [][]any{}
	for _, row := range rows {
		newRow := []any{}
		for _, col := range tablemeta.Edges.Columns {
			if v, ok := row[col.Name]; ok {
				newRow = append(newRow, util.ConvertAnyToType(v, col.Type))
			} else if v, ok := row[col.Nanoid]; ok {
				newRow = append(newRow, util.ConvertAnyToType(v, col.Type))
			} else {
				v, err = util.ConvertStringToType("", col.Type)
				if err != nil {
					return ent.Rollback(tx, err)
				}
				newRow = append(newRow, v)
			}
		}
		importRows = append(importRows, newRow)
	}

	rowCreates := []*ent.TableRowCreate{}
	for _, row := range importRows {
		cells := []*schema.CellValue{}
		for _, cell := range row {
			cells = append(cells, &schema.CellValue{Value: cell})
		}
		rowCreates = append(rowCreates, tx.TableRow.Create().SetCells(cells).SetTablemeta(tablemeta))
	}
	err = tx.TableRow.CreateBulk(rowCreates...).Exec(ctx)
	if err != nil {
		return ent.Rollback(tx, err)
	}
	return tx.Commit()
}
