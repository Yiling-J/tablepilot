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
	"github.com/Yiling-J/tablepilot/services/table/source/huggingface"
	"github.com/Yiling-J/tablepilot/services/table/util"
	"github.com/spf13/cast"

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
	Create(ctx context.Context, req *TableGenRequest) (string, error)
	Update(ctx context.Context, table string, req *TableGenRequest) (string, error)
	ListTables(ctx context.Context) (*ListTablesResponse, error)
	GetTableDetail(ctx context.Context, table string) (*TableInfo, error)
	Genetate(ctx context.Context, params GenerateRowsRequest) (RowsGenerator, error)
	Rows(ctx context.Context, table string) (*Rows, error)
	Truncate(ctx context.Context, table string) (int, error)
	Delete(ctx context.Context, table string) (int, error)
	Import(ctx context.Context, table string, reader io.Reader) (string, error)
	CreateRows(ctx context.Context, table string, rows []map[string]any) error
	SharedSources(ctx context.Context) []*SharedSource
}

type TableServiceImpl struct {
	config        *config.Config
	db            *ent.Client
	ai            ai.AiService
	sharedSources []*SharedSource
	logger        *zap.SugaredLogger
}

func NewTableService(config *config.Config, db *ent.Client, ai ai.AiService, logger *zap.SugaredLogger) (*TableServiceImpl, error) {
	ts := &TableServiceImpl{
		config:        config,
		db:            db,
		ai:            ai,
		sharedSources: []*SharedSource{},
		logger:        logger.With("service", "table"),
	}
	for _, s := range config.Sources {
		name := cast.ToString(s["name"])
		if name != "" {
			bs, err := json.Marshal(s)
			if err != nil {
				return nil, err
			}
			so, err := source.ValidateSource(context.Background(), json.RawMessage(bs), db)
			if err != nil {
				return nil, err
			}
			ss := &SharedSource{Name: name}
			switch st := so.(type) {
			case *source.CsvSource:
				columns, err := st.GetColumns(context.Background(), ts.logger, config.Common.SourceDataDir)
				if err != nil {
					return nil, err
				}
				ss.Columns = columns
			}
			ss.Data = json.RawMessage(bs)
			ts.sharedSources = append(ts.sharedSources, ss)
		}
	}
	return ts, nil
}

type TableGenColumnSchema struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
}

func (t *TableServiceImpl) Create(ctx context.Context, req *TableGenRequest) (string, error) {
	t.logger.Debugw("creating table", "name", req.Name)
	tx, err := t.db.Tx(ctx)
	if err != nil {
		return "", fmt.Errorf("starting a transaction: %w", err)
	}
	if req.Name == "" {
		return "", ent.Rollback(tx, errors.New("table name is empty"))
	}
	sources := map[string]json.RawMessage{}
	if len(req.Sources) > 0 {
		for _, raw := range req.Sources {
			vs, err := source.ValidateSource(ctx, raw, tx.Client())
			if err != nil {
				return "", ent.Rollback(tx, err)
			}
			if req.APIRequest() {
				switch st := vs.(type) {
				case *source.CsvSource:
					if len(st.Paths) > 0 {
						return "", errors.New("paths field for csv source is only allowed in CLI")
					}
				case *source.ListSource:
					if st.File != "" {
						return "", errors.New("file field for list source is only allowed in CLI")
					}
				}
			}
			bs, err := json.Marshal(vs)
			if err != nil {
				return "", ent.Rollback(tx, err)
			}
			sources[gjson.GetBytes(raw, "name").String()] = bs
		}
	}
	if len(t.sharedSources) > 0 {
		for _, so := range t.sharedSources {
			if _, ok := sources[so.Name]; !ok {
				vs, err := source.ValidateSource(ctx, so.Data, tx.Client())
				if err != nil {
					return "", ent.Rollback(tx, err)
				}
				bs, err := json.Marshal(vs)
				if err != nil {
					return "", ent.Rollback(tx, err)
				}
				sources[so.Name] = bs
			}
		}
	}

	table, err := tx.TableMeta.Create().SetName(req.Name).SetDescription(req.Description).SetModel(req.Model).SetSources(sources).Save(ctx)
	if err != nil {
		return "", ent.Rollback(tx, err)
	}

	// validate linked column column/context_columns exists
	err = validateLinkedColumnInfo(ctx, tx, req.Columns, sources)
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
			if _, ok := sources[col.Source]; ok {
				cc.SetSource(col.Source).SetRandom(col.Random).
					SetReplacement(col.Replacement).
					SetRepeat(col.Repeat).SetLinkedColumn(col.LinkedColumn).
					SetLinkedContextColumns(col.LinkedContextColumns)
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

func (t *TableServiceImpl) Update(ctx context.Context, table string, req *TableGenRequest) (string, error) {
	t.logger.Debugw("updating table", "table", table)
	tx, err := t.db.Tx(ctx)
	if err != nil {
		return "", fmt.Errorf("starting a transaction: %w", err)
	}
	if req.Name == "" {
		return "", ent.Rollback(tx, errors.New("table name is empty"))
	}
	if table == "" {
		table = req.Name
	}
	sources := map[string]json.RawMessage{}
	if len(req.Sources) > 0 {
		for _, raw := range req.Sources {
			vs, err := source.ValidateSource(ctx, raw, tx.Client())
			if err != nil {
				return "", ent.Rollback(tx, err)
			}
			if req.APIRequest() {
				switch st := vs.(type) {
				case *source.CsvSource:
					if len(st.Paths) > 0 {
						return "", errors.New("paths field for csv source is only allowed in CLI")
					}
				case *source.ListSource:
					if st.File != "" {
						return "", errors.New("file field for list source is only allowed in CLI")
					}
				}
			}
			bs, err := json.Marshal(vs)
			if err != nil {
				return "", ent.Rollback(tx, err)
			}
			sources[gjson.GetBytes(raw, "name").String()] = bs
		}
	}
	if len(t.sharedSources) > 0 {
		for _, so := range t.sharedSources {
			if _, ok := sources[so.Name]; !ok {
				vs, err := source.ValidateSource(ctx, so.Data, tx.Client())
				if err != nil {
					return "", ent.Rollback(tx, err)
				}
				bs, err := json.Marshal(vs)
				if err != nil {
					return "", ent.Rollback(tx, err)
				}
				sources[so.Name] = bs
			}
		}
	}

	dbtable, err := tx.TableMeta.Query().Where(tablemeta.Or(
		tablemeta.Nanoid(table),
		tablemeta.Name(table),
	)).WithColumns().Only(ctx)
	if err != nil {
		return "", ent.Rollback(tx, err)
	}

	err = dbtable.Update().SetName(req.Name).SetDescription(req.Description).SetModel(req.Model).SetSources(sources).Exec(ctx)
	if err != nil {
		return "", ent.Rollback(tx, err)
	}

	// validate linked column column/context_columns exists
	err = validateLinkedColumnInfo(ctx, tx, req.Columns, sources)
	if err != nil {
		return "", ent.Rollback(tx, err)
	}

	reqColumns := map[string]TableGenColumn{}
	for _, col := range req.Columns {
		reqColumns[col.Name] = col
	}

	removed := map[int]bool{}
	upserts := []TableGenColumn{}
	exists := map[string]string{}
	for i, col := range dbtable.Edges.Columns {
		if v, ok := reqColumns[col.Name]; !ok {
			removed[i] = true
			err = tx.TableColumn.DeleteOneID(col.ID).Exec(ctx)
			if err != nil {
				return "", ent.Rollback(tx, err)
			}
		} else {
			upserts = append(upserts, v)
			exists[col.Name] = col.Nanoid
		}
	}

	// append new columns and prepare zero values for existing rows
	zeros := []any{}
	for _, col := range req.Columns {
		if col.Name == "" {
			continue
		}
		if _, ok := exists[col.Name]; ok {
			continue
		}
		upserts = append(upserts, col)
		zv, err := util.ZeroValue(tablecolumn.Type(col.Type))
		if err != nil {
			return "", ent.Rollback(tx, err)
		}
		zeros = append(zeros, zv)
	}

	columnCreates := []*ent.TableColumnCreate{}
	for _, col := range upserts {
		cr := tx.TableColumn.Create()
		if nid, ok := exists[col.Name]; ok {
			cr.SetNanoid(nid)
		}
		cc := cr.SetTablemeta(dbtable).
			SetFillMode(tablecolumn.FillMode(col.FillMode)).
			SetName(col.Name).
			SetDescription(col.Description).
			SetType(tablecolumn.Type(col.Type)).
			SetContextLength(col.ContextLength)

		if col.Source != "" && col.FillMode == "pick" {
			if _, ok := sources[col.Source]; ok {
				cc.SetSource(col.Source).SetRandom(col.Random).
					SetReplacement(col.Replacement).
					SetRepeat(col.Repeat).SetLinkedColumn(col.LinkedColumn).
					SetLinkedContextColumns(col.LinkedContextColumns)
			} else {
				return "", ent.Rollback(tx, fmt.Errorf("source %s not dound", col.Source))
			}
		}
		columnCreates = append(columnCreates, cc)
		t.logger.Debugw(
			"upserting column", "name", col.Name, "description", col.Description,
			"fill_mode", col.FillMode, "type", col.Type, "source", col.Source,
		)
	}

	err = tx.TableColumn.CreateBulk(columnCreates...).OnConflictColumns(tablecolumn.FieldNanoid).UpdateNewValues().Exec(ctx)
	if err != nil {
		return "", ent.Rollback(tx, err)
	}

	// update rows if exists
	rows, err := dbtable.QueryRows().All(ctx)
	if err != nil {
		return "", ent.Rollback(tx, err)
	}
	updates := []*ent.TableRowCreate{}
	for _, row := range rows {
		cells := []*schema.CellValue{}
		for i, cell := range row.Cells {
			if !removed[i] {
				cells = append(cells, cell)
			}
		}
		for _, v := range zeros {
			cells = append(cells, &schema.CellValue{Value: v})
		}
		updates = append(updates, tx.TableRow.Create().SetTablemeta(dbtable).SetCells(cells).SetNanoid(row.Nanoid))
	}
	if len(updates) > 0 {
		err = tx.TableRow.CreateBulk(updates...).OnConflictColumns(tablerow.FieldNanoid).UpdateNewValues().Exec(ctx)
		if err != nil {
			return "", ent.Rollback(tx, err)
		}
	}

	t.logger.Debug("finish updating table")
	return dbtable.Nanoid, tx.Commit()
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
	params.sharedSources = map[string]json.RawMessage{}
	for _, so := range t.sharedSources {
		params.sharedSources[so.Name] = so.Data
	}
	params.sourceDataDir = t.config.Common.SourceDataDir
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

	var tableColumns []*ent.TableColumn
	if tablemeta != nil {
		cm := map[string]int{}
		for i, col := range columns {
			cm[col] = i
		}
		tableColumns = tablemeta.Edges.Columns
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
		tableColumns, err = tx.TableColumn.CreateBulk(columnCreates...).Save(ctx)
		if err != nil {
			return "", ent.Rollback(tx, err)
		}
	}

	schemaRows := [][]*schema.CellValue{}
	for _, row := range importRows {
		cells := []*schema.CellValue{}
		for _, cell := range row {
			cells = append(cells, &schema.CellValue{Value: cell})
		}
		schemaRows = append(schemaRows, cells)
	}
	for i, col := range tableColumns {
		if len(col.Source) == 0 {
			continue
		}
		rs, ok := tablemeta.Sources[col.Source]
		if !ok {
			return "", ent.Rollback(tx, errors.New("source not found"))
		}
		so, err := t.parseLinkedSource(ctx, tx.Client(), rs)
		if err != nil {
			return "", ent.Rollback(tx, err)
		}

		// assign context value to cell if column fill type is pick
		switch ts := so.(type) {
		case *source.ListSource:
			// there is no context value if source type is list
		case *source.LinkedSource:
			ts.Range(func(row *ent.TableRow) bool {
				v := ts.GetLinkedCellValue(row, col.LinkedColumn, col.LinkedContextColumns)
				for _, irow := range schemaRows {
					if irow[i].Value == v.Value {
						irow[i].ContextValue = v.ContextValue
					}
				}
				return true
			})
		case *source.CsvSource:
			err = ts.Range(func(row []any) bool {
				v := ts.GetLinkedCellValue(row, col.LinkedColumn, col.LinkedContextColumns)
				for _, irow := range schemaRows {
					if irow[i].Value == v.Value {
						irow[i].ContextValue = v.ContextValue
					}
				}
				return true
			})
			if err != nil {
				return "", ent.Rollback(tx, err)
			}
		case *source.ParquetSource:
			err = ts.Range(ctx, func(row map[string]any) bool {
				v := ts.GetLinkedCellValue(row, col.LinkedColumn, col.LinkedContextColumns)
				for _, irow := range schemaRows {
					if irow[i].Value == v.Value {
						irow[i].ContextValue = v.ContextValue
					}
				}
				return true
			})
			if err != nil {
				return "", ent.Rollback(tx, err)
			}
		}
	}

	rowCreates := []*ent.TableRowCreate{}
	for _, row := range schemaRows {
		rowCreates = append(rowCreates, tx.TableRow.Create().SetCells(row).SetTablemeta(tablemeta))
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

func (t *TableServiceImpl) SharedSources(ctx context.Context) []*SharedSource {
	return t.sharedSources
}

func (t *TableServiceImpl) parseLinkedSource(ctx context.Context, db *ent.Client, raw json.RawMessage) (source.Source, error) {
	sourceDataDir := t.config.Common.SourceDataDir
	var so source.Source
	sourceType := gjson.GetBytes(raw, "type").String()
	switch sourceType {
	case "list":
		var ls source.ListSource
		err := json.Unmarshal(raw, &ls)
		if err != nil {
			return nil, err
		}
		err = ls.Init(ctx, sourceDataDir)
		if err != nil {
			return nil, err
		}
		so = &ls
	case "linked":
		var ls source.LinkedSource
		err := json.Unmarshal(raw, &ls)
		if err != nil {
			return nil, err
		}
		err = ls.Init(ctx, db)
		if err != nil {
			return nil, err
		}
		so = &ls
	case "csv":
		var ls source.CsvSource
		err := json.Unmarshal(raw, &ls)
		if err != nil {
			return nil, err
		}
		err = ls.Init(ctx, t.logger, sourceDataDir)
		if err != nil {
			return nil, err
		}
		so = &ls
	case "parquet":
		var ls source.ParquetSource
		err := json.Unmarshal(raw, &ls)
		if err != nil {
			return nil, err
		}
		var client huggingface.Client
		if ls.Huggingface != nil {
			if ls.Huggingface.Dataset == "" {
				return nil, errors.New("dataset is empty")
			}
			if ls.Huggingface.Config == "" {
				ls.Huggingface.Config = "default"
			}
			if ls.Huggingface.Split == "" {
				ls.Huggingface.Split = "train"
			}
			client = huggingface.NewClient(
				ls.Huggingface.Dataset, ls.Huggingface.Config, ls.Huggingface.Split,
				t.logger,
			)
		}
		err = ls.Init(ctx, client, t.logger, sourceDataDir)
		if err != nil {
			return nil, err
		}
		so = &ls
	default:
		return nil, fmt.Errorf("unknown source type %s", sourceType)
	}
	return so, nil
}
