package table

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/Yiling-J/tablepilot/config"
	"github.com/Yiling-J/tablepilot/ent"
	"github.com/Yiling-J/tablepilot/ent/dataset"
	"github.com/Yiling-J/tablepilot/ent/schema"
	"github.com/Yiling-J/tablepilot/ent/tablecolumn"
	"github.com/Yiling-J/tablepilot/ent/tablemeta"
	"github.com/Yiling-J/tablepilot/ent/tablerow"
	"github.com/Yiling-J/tablepilot/services/ai"
	"github.com/Yiling-J/tablepilot/services/ai/client"
	"github.com/Yiling-J/tablepilot/services/ai/promptbuilder"
	dataset_service "github.com/Yiling-J/tablepilot/services/dataset"
	"github.com/Yiling-J/tablepilot/services/source"
	"github.com/Yiling-J/tablepilot/services/source/csvindexer"
	"github.com/Yiling-J/tablepilot/services/table/util"
	"github.com/Yiling-J/tablepilot/utils"
	"github.com/spf13/cast"
	orderedmap "github.com/wk8/go-ordered-map/v2"

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
	Validate(ctx context.Context, req *TableGenRequest) error
	Create(ctx context.Context, req *TableGenRequest) (string, error)
	Update(ctx context.Context, table string, req *TableGenRequest) (string, error)
	ListTables(ctx context.Context) (*ListTablesResponse, error)
	GetTableDetail(ctx context.Context, table string) (*TableInfo, error)
	Genetate(ctx context.Context, params GenerateRowsRequest) (RowsGenerator, error)
	Rows(ctx context.Context, table string) (*Rows, error)
	Truncate(ctx context.Context, table string) (int, error)
	Delete(ctx context.Context, table string) (int, error)
	Import(ctx context.Context, request ImportRequest) (string, error)
	ImportImage(ctx context.Context, request ImportRequest) (string, error)
	CreateRows(ctx context.Context, table string, rows []map[string]any) error
	GetTableSchema(ctx context.Context, table string) (*TableGenRequest, error)
	GenerateBuilderTables(ctx context.Context, prompt string, params ModelParams) ([]BuilderTable, error)
	PolishBuilderTables(ctx context.Context, tables []BuilderTable, prompt string, params ModelParams) ([]BuilderTable, error)
	BuildTable(ctx context.Context, name, description string, depends []string, exists []*TableInfo, params ModelParams) (*TableGenRequest, error)
	PolishBuilderTable(ctx context.Context, table *TableGenRequest, prompt string, exists []*TableInfo, params ModelParams) (*TableGenRequest, error)
	CSV(ctx context.Context, table string) ([]byte, error)
	CreateColumn(ctx context.Context, table string, column TableGenColumn) (string, error)
	DeleteColumn(ctx context.Context, table string, column string) (string, error)
}

type TableServiceImpl struct {
	config  *config.Config
	db      *ent.Client
	ai      ai.AiService
	dataset dataset_service.DatasetService
	logger  *zap.SugaredLogger
}

func NewTableService(config *config.Config, db *ent.Client, ai ai.AiService, dataset dataset_service.DatasetService, logger *zap.SugaredLogger) (*TableServiceImpl, error) {
	ts := &TableServiceImpl{
		config:  config,
		db:      db,
		ai:      ai,
		dataset: dataset,
		logger:  logger.With("service", "table"),
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
		return "", fmt.Errorf("table.Create: starting a transaction: %w", err)
	}
	if req.Name == "" {
		return "", ent.Rollback(tx, fmt.Errorf("table.Create: table name is empty"))
	}

	table, err := tx.TableMeta.Create().SetName(req.Name).SetDescription(req.Description).SetModel(req.Model).Save(ctx)
	if err != nil {
		return "", ent.Rollback(tx, fmt.Errorf("table.Create: saving table metadata: %w", err))
	}

	// validate linked column column/context_columns exists
	err = validateLinkedColumnInfo(ctx, tx.Client(), req.Columns)
	if err != nil {
		return "", ent.Rollback(tx, fmt.Errorf("table.Create: validating linked column info: %w", err))
	}

	extra := 0
	columns := []*TableGenColumn{}
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
			return "", fmt.Errorf("table.Create: building columns prompt: %w", err)
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
			return "", fmt.Errorf("table.Create: generating columns with AI: %w", err)
		}
		t.logger.Debug("extra columns generated")
		extraColumns, err := util.TryDecodeJsonArray[*TableGenColumn](resp.Content)
		if err != nil && len(extraColumns) == 0 {
			return "", fmt.Errorf("table.Create: decoding generated columns: %w", err)
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

		if col.FillMode == "pick" {
			cc.SetSourceID(col.SourceID).SetSourceType(col.SourceType).SetRandom(col.Random).
				SetReplacement(col.Replacement).
				SetRepeat(col.Repeat).SetLinkedColumn(col.LinkedColumn).
				SetLinkedContextColumns(col.LinkedContextColumns).
				SetOptions(col.Options)
		}
		columnCreates = append(columnCreates, cc)
		t.logger.Debugw(
			"creating column", "name", col.Name, "description", col.Description,
			"fill_mode", col.FillMode, "type", col.Type, "source_id", col.SourceID,
			"source_type", col.SourceType,
		)
	}

	err = tx.TableColumn.CreateBulk(columnCreates...).Exec(ctx)
	if err != nil {
		return "", ent.Rollback(tx, fmt.Errorf("table.Create: creating columns: %w", err))
	}

	t.logger.Debug("finish creating table")
	return table.Nanoid, tx.Commit()
}

func (t *TableServiceImpl) Update(ctx context.Context, table string, req *TableGenRequest) (string, error) {
	t.logger.Debugw("updating table", "table", table)
	tx, err := t.db.Tx(ctx)
	if err != nil {
		return "", fmt.Errorf("table.Update: starting a transaction: %w", err)
	}
	if req.Name == "" {
		return "", ent.Rollback(tx, fmt.Errorf("table.Update: table name is empty"))
	}
	if table == "" {
		table = req.Name
	}
	dbtable, err := tx.TableMeta.Query().Where(tablemeta.Or(
		tablemeta.Nanoid(table),
		tablemeta.Name(table),
	)).WithColumns().Only(ctx)
	if err != nil {
		return "", ent.Rollback(tx, fmt.Errorf("table.Update: querying table metadata: %w", err))
	}

	err = dbtable.Update().SetName(req.Name).SetDescription(req.Description).SetModel(req.Model).Exec(ctx)
	if err != nil {
		return "", ent.Rollback(tx, fmt.Errorf("table.Update: updating table metadata: %w", err))
	}

	// validate linked column column/context_columns exists
	err = validateLinkedColumnInfo(ctx, tx.Client(), req.Columns)
	if err != nil {
		return "", ent.Rollback(tx, fmt.Errorf("table.Update: validating linked column info: %w", err))
	}

	reqColumns := map[string]*TableGenColumn{}
	for _, col := range req.Columns {
		reqColumns[col.Name] = col
	}

	removed := map[int]bool{}
	upserts := []*TableGenColumn{}
	exists := map[string]string{}
	for i, col := range dbtable.Edges.Columns {
		if v, ok := reqColumns[col.Name]; !ok {
			removed[i] = true
			err = tx.TableColumn.DeleteOneID(col.ID).Exec(ctx)
			if err != nil {
				return "", ent.Rollback(tx, fmt.Errorf("table.Update: deleting column: %w", err))
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
			return "", ent.Rollback(tx, fmt.Errorf("table.Update: getting zero value for column type: %w", err))
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

		if col.SourceID != "" && col.FillMode == "pick" {
			cc.SetSourceID(col.SourceID).SetSourceType(col.SourceType).SetRandom(col.Random).
				SetReplacement(col.Replacement).
				SetRepeat(col.Repeat).SetLinkedColumn(col.LinkedColumn).
				SetLinkedContextColumns(col.LinkedContextColumns)
		}
		columnCreates = append(columnCreates, cc)
		t.logger.Debugw(
			"upserting column", "name", col.Name, "description", col.Description,
			"fill_mode", col.FillMode, "type", col.Type, "source_id", col.SourceID,
			"source_type", col.SourceType,
		)
	}

	err = tx.TableColumn.CreateBulk(columnCreates...).OnConflictColumns(tablecolumn.FieldNanoid).UpdateNewValues().Exec(ctx)
	if err != nil {
		return "", ent.Rollback(tx, fmt.Errorf("table.Update: upserting columns: %w", err))
	}

	// update rows if exists
	rows, err := dbtable.QueryRows().All(ctx)
	if err != nil {
		return "", ent.Rollback(tx, fmt.Errorf("table.Update: querying rows: %w", err))
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
			return "", ent.Rollback(tx, fmt.Errorf("table.Update: updating rows: %w", err))
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
		return nil, fmt.Errorf("table.ListTables: querying tables: %w", err)
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
		return nil, fmt.Errorf("table.GetTableDetail: querying table: %w", err)
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
	params.sourceDataDir = t.config.Common.SourceDataDir
	generator, err := NewRowsGenerator(ctx, params, t.db, t.ai, t.logger)
	if err != nil {
		return nil, fmt.Errorf("table.Genetate: creating rows generator: %w", err)
	}
	return generator, nil
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
		return nil, fmt.Errorf("table.Rows: querying table: %w", err)
	}
	return &Rows{Columns: meta.Edges.Columns, Rows: meta.Edges.Rows}, nil
}

func cellString(v any) string {
	vs, err := cast.ToStringE(v)
	if err != nil {
		vb, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%+v", v)
		}
		return string(vb)
	}
	return vs
}

func (t *TableServiceImpl) CSV(ctx context.Context, table string) ([]byte, error) {
	rows, err := t.Rows(ctx, table)
	if err != nil {
		return nil, fmt.Errorf("table.CSV: querying rows: %w", err)
	}

	buffer := bytes.NewBuffer([]byte{})
	csvwriter := csv.NewWriter(buffer)
	columns := []string{}
	for _, col := range rows.Columns {
		columns = append(columns, col.Name)
	}
	err = csvwriter.Write(columns)
	if err != nil {
		return nil, fmt.Errorf("table.CSV: write headers to csv: %w", err)
	}
	data := [][]string{}
	for _, row := range rows.Rows {
		r := []string{}
		for _, v := range row.Cells {
			r = append(r, cellString(v.Value))
		}
		data = append(data, r)
	}
	err = csvwriter.WriteAll(data)
	if err != nil {
		return nil, fmt.Errorf("table.CSV: write data to csv: %w", err)
	}
	return buffer.Bytes(), nil
}

func (t *TableServiceImpl) Truncate(ctx context.Context, table string) (int, error) {
	meta, err := t.db.TableMeta.Query().Where(tablemeta.Or(
		tablemeta.Nanoid(table),
		tablemeta.Name(table),
	)).First(ctx)
	if err != nil {
		return 0, fmt.Errorf("table.Truncate: querying table: %w", err)
	}
	count, err := t.db.TableRow.Delete().Where(tablerow.HasTablemetaWith(tablemeta.Nanoid(meta.Nanoid))).Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("table.Truncate: deleting rows: %w", err)
	}
	return count, nil
}

func (t *TableServiceImpl) Delete(ctx context.Context, table string) (int, error) {
	count, err := t.db.TableMeta.Delete().Where(tablemeta.Or(
		tablemeta.Nanoid(table),
		tablemeta.Name(table),
	)).Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("table.Delete: deleting table: %w", err)
	}
	return count, nil
}

func (t *TableServiceImpl) ImportImage(ctx context.Context, request ImportRequest) (string, error) {
	builder := promptbuilder.NewNewImageToTableBuilder(request.Prompt)
	var genSchema *jsonschema.Schema
	tx, err := t.db.Tx(ctx)
	if err != nil {
		return "", fmt.Errorf("table.ImportImage: starting transaction: %w", err)
	}
	if len(request.Table) > 0 {
		table, err := tx.TableMeta.Query().WithColumns().Where(tablemeta.Or(
			tablemeta.Nanoid(request.Table),
			tablemeta.Name(request.Table),
		)).Only(ctx)
		if err != nil {
			return "", fmt.Errorf("table.ImportImage: get table: %w", err)
		}
		if request.Truncate {
			_, err = tx.TableRow.Delete().Where(tablerow.HasTablemetaWith(tablemeta.ID(table.ID))).Exec(ctx)
			if err != nil {
				return "", ent.Rollback(tx, fmt.Errorf("table.ImportImage: truncating existing table: %w", err))
			}
		}
		builder.ToTable(table)

		om := orderedmap.New[string, *jsonschema.Schema]()
		om.Set("__id__", &jsonschema.Schema{Type: "integer"})
		required := []string{"__id__"}
		for _, col := range table.Edges.Columns {
			if col.FillMode != tablecolumn.FillModeAi {
				continue
			}
			s := &jsonschema.Schema{
				Type: col.Type.String(),
			}
			if s.Type == "array" {
				s.Items = &jsonschema.Schema{Type: "string"}
			}
			om.Set(col.Nanoid, s)
			required = append(required, col.Nanoid)
		}
		rowsSchema := &jsonschema.Schema{
			Type: "array",
			Items: &jsonschema.Schema{
				Type:                 "object",
				Properties:           om,
				AdditionalProperties: jsonschema.FalseSchema,
				Required:             required,
			},
		}
		omw := orderedmap.New[string, *jsonschema.Schema]()
		omw.Set("data", rowsSchema)
		genSchema = &jsonschema.Schema{
			Type:                 "object",
			Properties:           omw,
			AdditionalProperties: jsonschema.FalseSchema,
		}
	} else {
		tables, err := tx.TableMeta.Query().All(ctx)
		if err != nil {
			return "", ent.Rollback(tx, fmt.Errorf("table.ImportImage: querying tables: %w", err))
		}
		tableNames := []string{}
		for _, t := range tables {
			tableNames = append(tableNames, t.Name)
		}
		builder.AddExistingTableNames(tableNames)
		var v ImageExtractionOutput
		reflector := jsonschema.Reflector{
			AllowAdditionalProperties: false,
			DoNotReference:            true,
		}
		genSchema = reflector.Reflect(v)
	}
	prompt, err := builder.Prompt()
	if err != nil {
		return "", ent.Rollback(tx, fmt.Errorf("table.ImportImage: building prompt: %w", err))
	}
	encoded, err := imageURLFromData(request.Data)
	if err != nil {
		return "", ent.Rollback(tx, fmt.Errorf("table.ImportImage: encoding image: %w", err))
	}
	input := client.UserMessageWithSingleImage(prompt, encoded)
	resp, err := t.ai.Chat(ctx, &client.ChatRequest{
		Temperature:     0.1,
		MaxOutputTokens: GENERATE_DATA_MAX_TOKENS,
		Messages:        []*client.Message{input},
		Model:           request.Model,
		Schema:          genSchema,
	})
	if err != nil {
		return "", ent.Rollback(tx, fmt.Errorf("table.ImportImage: AI chat request: %w", err))
	}

	creates := []*ent.TableRowCreate{}
	tableID := ""
	if len(request.Table) > 0 {
		table, err := tx.TableMeta.Query().WithColumns().Where(tablemeta.Or(
			tablemeta.Nanoid(request.Table),
			tablemeta.Name(request.Table),
		)).Only(ctx)
		tableID = table.Nanoid
		if err != nil {
			return "", ent.Rollback(tx, fmt.Errorf("table.ImportImage: get table: %w", err))
		}
		rows, err := util.TryDecodeJsonArray[map[string]any](
			gjson.Get(resp.Content, "data").String(),
		)
		// log error only because we need those successfully generated rows
		if err != nil {
			t.logger.Errorw("table.ImportImage: TryDecodeJsonArray error", "errpr", err)
		}

		generated := []map[string]*schema.CellValue{}
		for _, row := range rows {
			n := map[string]*schema.CellValue{}
			for k, v := range row {
				n[k] = &schema.CellValue{Value: v}
			}
			delete(row, "__id__")
			generated = append(generated, n)
		}
		indexer := util.NewColumnIndexer(table.Edges.Columns)
		for _, row := range generated {
			v, err := indexer.RowMapToSlice(row)
			if err != nil {
				return "", ent.Rollback(tx, fmt.Errorf("table.ImportImage: RowMapToSlice: %w", err))
			}
			creates = append(creates, tx.TableRow.Create().SetNanoid(
				cast.ToString(row["__id__"].Value),
			).SetCells(v).SetTablemeta(table))
		}
	} else {
		var generated ImageExtractionOutput
		err = json.Unmarshal([]byte(resp.Content), &generated)
		if err != nil {
			return "", ent.Rollback(tx, fmt.Errorf("table.ImportImage: unmarshaling AI response: %w", err))
		}
		name := generated.TableName
		if len(request.Name) > 0 {
			name = request.Name
		}
		tablemeta, err := tx.TableMeta.Create().SetName(name).SetDescription(generated.TableDescription).Save(ctx)
		if err != nil {
			return "", ent.Rollback(tx, fmt.Errorf("table.ImportImage: creating table metadata: %w", err))
		}
		tableID = tablemeta.Nanoid
		columnCreates := []*ent.TableColumnCreate{}
		for _, col := range generated.Columns {
			columnCreates = append(
				columnCreates, tx.TableColumn.Create().SetName(col.Name).SetType(tablecolumn.TypeString).SetDescription(col.Description).
					SetFillMode(tablecolumn.FillModeAi).SetTablemeta(tablemeta),
			)
		}
		err = tx.TableColumn.CreateBulk(columnCreates...).Exec(ctx)
		if err != nil {
			return "", ent.Rollback(tx, fmt.Errorf("table.ImportImage: creating columns: %w", err))
		}
		schemaRows := [][]*schema.CellValue{}
		for _, row := range generated.Rows {
			if len(row) != len(generated.Columns) {
				return "", ent.Rollback(tx, fmt.Errorf("table.ImportImage: inconsistent row data"))
			}
			cells := []*schema.CellValue{}
			for _, cell := range row {
				cells = append(cells, &schema.CellValue{Value: cell})
			}
			schemaRows = append(schemaRows, cells)
		}

		for _, row := range schemaRows {
			creates = append(creates, tx.TableRow.Create().SetCells(row).SetTablemeta(tablemeta))
		}
	}
	err = tx.TableRow.CreateBulk(creates...).Exec(ctx)
	if err != nil {
		return "", ent.Rollback(tx, fmt.Errorf("table.ImportImage: creating rows: %w", err))
	}
	return tableID, tx.Commit()
}

func (t *TableServiceImpl) Import(ctx context.Context, request ImportRequest) (string, error) {
	tx, err := t.db.Tx(ctx)
	if err != nil {
		return "", fmt.Errorf("table.Import: starting a transaction: %w", err)
	}
	var tm *ent.TableMeta
	if len(request.Table) > 0 {
		tm, err = tx.TableMeta.Query().WithColumns(func(tcq *ent.TableColumnQuery) {
			tcq.Order(ent.Asc(tablecolumn.FieldID))
		}).Where(
			tablemeta.Or(
				tablemeta.Nanoid(request.Table),
				tablemeta.Name(request.Table),
			),
		).Only(ctx)
		if err != nil {
			return "", ent.Rollback(tx, fmt.Errorf("table.Import: querying existing table: %w", err))
		}
		if request.Truncate {
			_, err = tx.TableRow.Delete().Where(tablerow.HasTablemetaWith(tablemeta.ID(tm.ID))).Exec(ctx)
			if err != nil {
				return "", ent.Rollback(tx, fmt.Errorf("table.Import: truncating existing table: %w", err))
			}
		}
	}
	cr := utils.NewCsvReader(request.Reader)
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
			return "", ent.Rollback(tx, fmt.Errorf("table.Import: reading CSV: %w", err))
		}
		if counter == 0 {
			columns = record
		} else {
			rows = append(rows, record)
		}
		counter += 1
	}

	var tableColumns []*ent.TableColumn
	if tm != nil {
		cm := map[string]int{}
		for i, col := range columns {
			cm[col] = i
		}
		tableColumns = tm.Edges.Columns
		for _, row := range rows {
			newRow := []any{}
			for _, col := range tm.Edges.Columns {
				if j, ok := cm[col.Name]; ok {
					v, err := util.ConvertStringToType(row[j], col.Type)
					if err != nil {
						return "", ent.Rollback(tx, fmt.Errorf("table.Import: converting string to type: %w", err))
					}
					newRow = append(newRow, v)
				} else {
					v, err := util.ConvertStringToType("", col.Type)
					if err != nil {
						return "", ent.Rollback(tx, fmt.Errorf("table.Import: converting empty string to type: %w", err))
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
		tableName := request.Name
		if len(tableName) == 0 {
			tableName = fmt.Sprintf("%s_%d", request.Filename, time.Now().Unix())
		}
		tm, err = tx.TableMeta.Create().SetName(tableName).Save(ctx)
		if err != nil {
			return "", ent.Rollback(tx, fmt.Errorf("table.Import: creating table metadata: %w", err))
		}
		columnCreates := []*ent.TableColumnCreate{}
		for _, col := range columns {
			columnCreates = append(
				columnCreates, tx.TableColumn.Create().SetName(col).SetType(tablecolumn.TypeString).
					SetFillMode(tablecolumn.FillModeAi).SetTablemeta(tm),
			)
		}
		tableColumns, err = tx.TableColumn.CreateBulk(columnCreates...).Save(ctx)
		if err != nil {
			return "", ent.Rollback(tx, fmt.Errorf("table.Import: creating columns: %w", err))
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
		if col.FillMode != tablecolumn.FillModePick {
			continue
		}

		// add context data to rows if source type is table or dataset-csv
		switch col.SourceType {
		case tablecolumn.SourceTypeTable:
			ts := &source.LinkedSource{Table: col.SourceID}
			err = ts.Init(ctx, tx.Client())
			if err != nil {
				return "", ent.Rollback(tx, fmt.Errorf("table.Import: init table source: %w", err))
			}
			ts.Range(func(row *ent.TableRow) bool {
				v := ts.GetLinkedCellValue(row, col.LinkedColumn, col.LinkedContextColumns)
				for _, irow := range schemaRows {
					if irow[i].Value == v.Value {
						irow[i].ContextValue = v.ContextValue
					}
				}
				return true
			})
		case tablecolumn.SourceTypeDataset:
			ds, err := tx.Dataset.Query().Where(dataset.Nanoid(col.SourceID)).Only(ctx)
			if err != nil {
				return "", ent.Rollback(tx, fmt.Errorf("table.Import: get column dataset: %w", err))
			}
			switch ds.Type {
			case dataset.TypeList:
			case dataset.TypeCsv:
				ts := &source.CsvSource{
					RandomCSV: &csvindexer.CSVIndexer{
						FS:         os.DirFS(ds.Path),
						CSVIndexer: ds.Indexer,
					},
				}
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
					return "", ent.Rollback(tx, fmt.Errorf("table.Import: ranging over CSV source: %w", err))
				}
			}
		case tablecolumn.SourceTypeOptions:
		}
	}

	rowCreates := []*ent.TableRowCreate{}
	for _, row := range schemaRows {
		rowCreates = append(rowCreates, tx.TableRow.Create().SetCells(row).SetTablemeta(tm))
	}
	err = tx.TableRow.CreateBulk(rowCreates...).Exec(ctx)
	if err != nil {
		return "", ent.Rollback(tx, fmt.Errorf("table.Import: creating rows: %w", err))
	}
	return tm.Nanoid, tx.Commit()
}

func (t *TableServiceImpl) CreateRows(ctx context.Context, table string, rows []map[string]any) error {
	tx, err := t.db.Tx(ctx)
	if err != nil {
		return fmt.Errorf("table.CreateRows: starting a transaction: %w", err)
	}
	tablemeta, err := tx.TableMeta.Query().WithColumns(func(tcq *ent.TableColumnQuery) {
		tcq.Order(ent.Asc(tablecolumn.FieldID))
	}).Where(
		tablemeta.Or(
			tablemeta.Nanoid(table),
			tablemeta.Name(table)),
	).First(ctx)
	if err != nil {
		return ent.Rollback(tx, fmt.Errorf("table.CreateRows: querying table metadata: %w", err))
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
					return ent.Rollback(tx, fmt.Errorf("table.CreateRows: converting empty string to type: %w", err))
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
		return ent.Rollback(tx, fmt.Errorf("table.CreateRows: creating rows: %w", err))
	}
	return tx.Commit()
}

func getSourceFromColumn(ctx context.Context, db *ent.Client, sourceDataDir string, column *ent.TableColumn) (source.Source, error) {
	var so source.Source
	switch column.SourceType {
	case tablecolumn.SourceTypeTable:
		ls := &source.LinkedSource{
			Table: column.SourceID,
		}
		err := ls.Init(ctx, db)
		if err != nil {
			return nil, fmt.Errorf("table.getSourceFromColumn: initializing linked source: %w", err)
		}
		so = ls
	case tablecolumn.SourceTypeDataset:
		ds, err := db.Dataset.Query().Where(dataset.Nanoid(column.SourceID)).Only(ctx)
		if err != nil {
			return nil, fmt.Errorf("table.getSourceFromColumn: get dataset: %w", err)
		}
		switch ds.Type {
		case dataset.TypeList:
			ls := &source.ListSource{Options: ds.Values}
			err = ls.Init(ctx, sourceDataDir)
			if err != nil {
				return nil, fmt.Errorf("table.parseLinkedSource: initializing list source: %w", err)
			}
			so = ls
		case dataset.TypeCsv:
			ls := &source.CsvSource{
				RandomCSV: &csvindexer.CSVIndexer{
					FS:         os.DirFS(ds.Path),
					CSVIndexer: ds.Indexer,
				},
			}
			so = ls
		}
	case tablecolumn.SourceTypeOptions:
		ls := &source.ListSource{Options: column.Options}
		err := ls.Init(ctx, sourceDataDir)
		if err != nil {
			return nil, fmt.Errorf("table.parseLinkedSource: initializing options source: %w", err)
		}
		so = ls
	}
	return so, nil
}

func (t *TableServiceImpl) GetTableSchema(ctx context.Context, table string) (*TableGenRequest, error) {
	meta, err := t.db.TableMeta.Query().WithColumns(func(tcq *ent.TableColumnQuery) {
		tcq.Order(ent.Asc(tablecolumn.FieldID))
	}).Where(tablemeta.Or(
		tablemeta.Nanoid(table),
		tablemeta.Name(table),
	)).First(ctx)
	if err != nil {
		return nil, fmt.Errorf("table.GetTableSchema: querying table metadata: %w", err)
	}
	schema := &TableGenRequest{
		Name:        meta.Name,
		Description: meta.Description,
	}
	columns := []*TableGenColumn{}
	for _, col := range meta.Edges.Columns {
		columns = append(columns, &TableGenColumn{
			Name:                 col.Name,
			Description:          col.Description,
			Type:                 col.Type.String(),
			FillMode:             col.FillMode.String(),
			Random:               col.Random,
			Replacement:          col.Replacement,
			Repeat:               col.Repeat,
			ContextLength:        col.ContextLength,
			LinkedColumn:         col.LinkedColumn,
			LinkedContextColumns: col.LinkedContextColumns,
			SourceID:             col.SourceID,
			SourceType:           col.SourceType,
			Options:              col.Options,
		})
	}
	schema.Columns = columns

	return schema, nil
}

func (t *TableServiceImpl) Validate(ctx context.Context, req *TableGenRequest) error {
	if len(req.Columns) == 0 {
		return fmt.Errorf("table.Validate: columns should not be empty")
	}
	return validateLinkedColumnInfo(ctx, t.db, req.Columns)
}

func (t *TableServiceImpl) CreateColumn(ctx context.Context, table string, column TableGenColumn) (string, error) {
	t.logger.Debugw("creating column", "column", column.Name)
	tx, err := t.db.Tx(ctx)
	if err != nil {
		return "", fmt.Errorf("table.CreateColumn: starting a transaction: %w", err)
	}
	tb, err := tx.TableMeta.Query().Where(tablemeta.Or(
		tablemeta.Nanoid(table),
		tablemeta.Name(table),
	)).Only(ctx)
	if err != nil {
		return "", fmt.Errorf("table.CreateColumn: get table: %w", err)
	}

	// validate linked column column/context_columns exists
	err = validateLinkedColumnInfo(ctx, tx.Client(), []*TableGenColumn{&column})
	if err != nil {
		return "", ent.Rollback(tx, fmt.Errorf("table.CreateColumn: validating linked column info: %w", err))
	}

	cc := tx.TableColumn.Create().
		SetTablemeta(tb).
		SetFillMode(tablecolumn.FillMode(column.FillMode)).
		SetName(column.Name).
		SetDescription(column.Description).
		SetType(tablecolumn.Type(column.Type)).
		SetContextLength(column.ContextLength)

	if column.SourceID != "" && column.FillMode == "pick" {
		cc.SetSourceID(column.SourceID).SetSourceType(column.SourceType).SetRandom(column.Random).
			SetReplacement(column.Replacement).
			SetRepeat(column.Repeat).SetLinkedColumn(column.LinkedColumn).
			SetLinkedContextColumns(column.LinkedContextColumns)
	}
	t.logger.Debugw(
		"creating column", "name", column.Name, "description", column.Description,
		"fill_mode", column.FillMode, "type", column.Type, "source_id", column.SourceID,
		"source_type", column.SourceType,
	)

	err = cc.Exec(ctx)
	if err != nil {
		return "", ent.Rollback(tx, fmt.Errorf("table.CreateColumn: creating column: %w", err))
	}

	// update rows if exists
	rows, err := tb.QueryRows().All(ctx)
	if err != nil {
		return "", ent.Rollback(tx, fmt.Errorf("table.CreateColumn: querying rows: %w", err))
	}
	updates := []*ent.TableRowCreate{}
	zv, err := util.ZeroValue(tablecolumn.Type(column.Type))
	if err != nil {
		return "", ent.Rollback(tx, fmt.Errorf("table.Update: getting zero value for column type: %w", err))
	}
	for _, row := range rows {
		row.Cells = append(row.Cells, &schema.CellValue{Value: zv})
		updates = append(updates, tx.TableRow.Create().SetTablemeta(tb).SetCells(row.Cells).SetNanoid(row.Nanoid))
	}
	if len(updates) > 0 {
		err = tx.TableRow.CreateBulk(updates...).OnConflictColumns(tablerow.FieldNanoid).UpdateNewValues().Exec(ctx)
		if err != nil {
			return "", ent.Rollback(tx, fmt.Errorf("table.Update: updating rows: %w", err))
		}
	}

	t.logger.Debug("finish creating column")
	return tb.Nanoid, tx.Commit()
}

func (t *TableServiceImpl) DeleteColumn(ctx context.Context, table string, column string) (string, error) {
	t.logger.Debugw("deleting column", "column", column)
	tx, err := t.db.Tx(ctx)
	if err != nil {
		return "", fmt.Errorf("table.DeleteColumn: starting a transaction: %w", err)
	}
	tb, err := tx.TableMeta.Query().Where(tablemeta.Or(
		tablemeta.Nanoid(table),
		tablemeta.Name(table),
	)).WithColumns(func(tcq *ent.TableColumnQuery) {
		tcq.Order(ent.Asc(tablecolumn.FieldID))
	}).Only(ctx)
	if err != nil {
		return "", ent.Rollback(tx, fmt.Errorf("table.DeleteColumn: get table: %w", err))
	}
	removeIndex := 0
	removeId := 0
	for i, col := range tb.Edges.Columns {
		if col.Nanoid == column || col.Name == column {
			removeIndex = i
			removeId = col.ID
			break
		}
	}
	if removeId == 0 {
		return "", nil
	}
	_, err = tx.TableColumn.Delete().Where(tablecolumn.ID(removeId)).Exec(ctx)
	if err != nil {
		return "", ent.Rollback(tx, fmt.Errorf("table.DeleteColumn: delete column: %w", err))
	}
	rows, err := tb.QueryRows().All(ctx)
	if err != nil {
		return "", ent.Rollback(tx, fmt.Errorf("table.DeleteColumn: querying rows: %w", err))
	}
	updates := []*ent.TableRowCreate{}
	for _, row := range rows {
		cells := []*schema.CellValue{}
		for i, cell := range row.Cells {
			if i != removeIndex {
				cells = append(cells, cell)
			}
		}
		updates = append(updates, tx.TableRow.Create().SetTablemeta(tb).SetCells(cells).SetNanoid(row.Nanoid))
	}
	if len(updates) > 0 {
		err = tx.TableRow.CreateBulk(updates...).OnConflictColumns(tablerow.FieldNanoid).UpdateNewValues().Exec(ctx)
		if err != nil {
			return "", ent.Rollback(tx, fmt.Errorf("table.DeleteColumn: updating rows: %w", err))
		}
	}
	return "", tx.Commit()
}
