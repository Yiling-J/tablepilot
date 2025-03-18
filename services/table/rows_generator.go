package table

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"

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
	"github.com/spf13/cast"

	"github.com/invopop/jsonschema"
	"github.com/tidwall/gjson"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"go.uber.org/zap"
)

type RowsGenerator interface {
	Next(ctx context.Context) ([]map[string]*schema.CellValue, error)
	Table() *ent.TableMeta
}

type AIRowsGenerator struct {
	db             *ent.Client
	ai             ai.AiService
	logger         *zap.SugaredLogger
	table          *ent.TableMeta
	missingColumns []*ent.TableColumn
	contextColumns []*ent.TableColumn
	indexerMap     map[string]*source.Indexer
	sourceMap      map[string]source.Source
	generated      []map[string]*schema.CellValue
	contextLength  int
	saveTo         string
	temperature    float64
	model          string

	total     int
	batchSize int
	current   int
	offset    int

	rows    []map[string]*schema.CellValue
	builder *promptbuilder.RowsBuilder

	autofill AutofillRequest
}

func NewRowsGenerator(ctx context.Context, params GenerateRowsRequest, db *ent.Client, ai ai.AiService, logger *zap.SugaredLogger) (*AIRowsGenerator, error) {
	generator := &AIRowsGenerator{
		logger: logger,
		db:     db,
		ai:     ai,

		total:       params.Count,
		batchSize:   params.Batch,
		indexerMap:  make(map[string]*source.Indexer),
		sourceMap:   make(map[string]source.Source),
		saveTo:      params.SaveTo,
		temperature: params.Temperature,
		model:       params.Model,
		autofill:    params.Autofill,
		offset:      params.Autofill.Offset,
	}
	meta, err := db.TableMeta.Query().WithColumns(func(tcq *ent.TableColumnQuery) {
		tcq.Order(ent.Asc(tablecolumn.FieldID))
	}).Where(tablemeta.Or(
		tablemeta.Nanoid(params.Table),
		tablemeta.Name(params.Table),
	)).First(ctx)
	if err != nil {
		return nil, err
	}
	generator.table = meta
	// in autofill mode, is ContextColumns is empty, add all other columns as context columns
	if generator.autofill.Enable && len(generator.autofill.ContextColumns) == 0 {
		for _, c := range meta.Edges.Columns {
			if !slices.Contains(params.Autofill.Columns, c.Name) && !slices.Contains(params.Autofill.Columns, c.Nanoid) {
				generator.autofill.ContextColumns = append(generator.autofill.ContextColumns, c.Nanoid)
			}
		}
	}

	for _, c := range meta.Edges.Columns {
		if params.Autofill.Enable {
			if slices.Contains(generator.autofill.ContextColumns, c.Name) || slices.Contains(generator.autofill.ContextColumns, c.Nanoid) {
				generator.contextColumns = append(generator.contextColumns, c)
			}
		} else {
			generator.contextColumns = append(generator.contextColumns, c)
		}
		// in autofill mode, only autofill required columns
		if params.Autofill.Enable && !slices.Contains(generator.autofill.Columns, c.Name) && !slices.Contains(generator.autofill.Columns, c.Nanoid) {
			continue
		}
		if c.ContextLength > generator.contextLength {
			generator.contextLength = c.ContextLength
		}
		if c.FillMode == tablecolumn.FillModePick {
			if len(c.Source) == 0 {
				return nil, errors.New("invalid source")
			}
			idx, err := generator.columnSourceIndexer(ctx, meta.Sources[c.Source], c)
			if err != nil {
				return nil, err
			}
			generator.indexerMap[c.Nanoid] = idx
			continue
		}
		generator.missingColumns = append(generator.missingColumns, c)
	}
	return generator, nil
}

func (g *AIRowsGenerator) newBatch(ctx context.Context, batch int) error {
	g.builder = promptbuilder.NewRowsBuilder(batch)
	g.builder.AddDescription(g.table.Description)
	g.rows = g.rows[:0]
	return g.prepareContextRows(ctx)
}

func (g *AIRowsGenerator) prepareContextRows(ctx context.Context) error {
	// get required rows from previous generated results and database
	contextRows := []map[string]*schema.CellValue{}
	if g.contextLength > 0 {
		remain := g.contextLength
		for i := range g.generated {
			contextRows = append(contextRows, g.generated[len(g.generated)-1-i])
			remain -= 1
			if remain == 0 {
				break
			}
		}
		if remain > 0 {
			rows, err := g.table.QueryRows().Order(
				ent.Desc(tablerow.FieldID),
			).Limit(remain).All(ctx)
			if err != nil {
				return err
			}
			for _, row := range rows {
				m := map[string]*schema.CellValue{}
				for i, col := range g.table.Edges.Columns {
					m[col.Nanoid] = row.Cells[i]
				}
				contextRows = append(contextRows, m)
			}
		}
	}

	for _, col := range g.contextColumns {
		if col.ContextLength > 0 {
			values := []any{}
			for i := 0; i < col.ContextLength; i++ {
				if i >= len(contextRows) {
					break
				}
				cv := contextRows[i][col.Nanoid]
				if cv.ContextValue != nil {
					values = append(values, cv.ContextValue)
				} else {
					values = append(values, cv.Value)
				}
			}
			err := g.builder.AddColumnContextData(col.Nanoid, values)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *AIRowsGenerator) prepareRows(ctx context.Context, batch int) error {
	if g.autofill.Enable {
		rows, err := g.table.QueryRows().Order(
			ent.Asc(tablerow.FieldID),
		).Limit(batch).Offset(g.offset).All(ctx)
		if err != nil {
			return err
		}
		g.offset += len(rows)
		for _, dbrow := range rows {
			row := map[string]*schema.CellValue{}
			for i, col := range g.table.Edges.Columns {
				row[col.Nanoid] = dbrow.Cells[i]
			}
			row["id"] = &schema.CellValue{Value: dbrow.Nanoid}
			g.rows = append(g.rows, row)
		}
	} else {
		for n := 0; n < batch; n++ {
			row := map[string]*schema.CellValue{}
			for _, col := range g.table.Edges.Columns {
				idx, ok := g.indexerMap[col.Nanoid]
				if ok {
					v, err := idx.Next(ctx)
					if err != nil {
						return err
					}
					row[col.Nanoid] = v
				}
			}
			if len(row) > 0 {
				row["id"] = &schema.CellValue{Value: n}
				g.rows = append(g.rows, row)
			}
		}
	}
	return nil
}

func (g *AIRowsGenerator) chat(ctx context.Context) (*client.ChatResponse, error) {
	om := orderedmap.New[string, *jsonschema.Schema]()
	if g.autofill.Enable {
		// row nanoid
		om.Set("id", &jsonschema.Schema{Type: "string"})
	} else {
		om.Set("id", &jsonschema.Schema{Type: "integer"})
	}
	required := []string{"id"}
	for _, col := range g.missingColumns {
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
	prompt, err := g.builder.Prompt()
	if err != nil {
		return nil, err
	}
	input := client.UserMessage(prompt)
	temperature := g.temperature
	if temperature < 0 {
		temperature = GENERATE_DATA_TEMPERATURE
	}
	model := g.model
	if model == "" {
		model = g.table.Model
	}
	return g.ai.Chat(ctx, &client.ChatRequest{
		Temperature:     temperature,
		MaxOutputTokens: GENERATE_DATA_MAX_TOKENS,
		Messages:        []*client.Message{input},
		Model:           model,
		Schema: &jsonschema.Schema{
			Type:                 "object",
			Properties:           omw,
			AdditionalProperties: jsonschema.FalseSchema,
		},
	})
}

func (g *AIRowsGenerator) columnSourceIndexer(ctx context.Context, raw json.RawMessage, column *ent.TableColumn) (*source.Indexer, error) {
	if so, ok := g.sourceMap[column.Source]; ok {
		return source.NewIndexer(so, column), nil
	}
	var so source.Source
	sourceType := gjson.GetBytes(raw, "type").String()
	switch sourceType {
	case "list":
		var ls source.ListSource
		err := json.Unmarshal(raw, &ls)
		if err != nil {
			return nil, err
		}
		err = ls.Init(ctx)
		if err != nil {
			return nil, err
		}
		so = &ls
	case "ai":
		var ls source.AISource
		err := json.Unmarshal(raw, &ls)
		if err != nil {
			return nil, err
		}
		err = ls.Init(ctx, g.ai, column, g.model)
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
		err = ls.Init(ctx, g.db)
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
		err = ls.Init(ctx)
		if err != nil {
			return nil, err
		}
		so = &ls
	default:
		return nil, fmt.Errorf("unknow source type %s", sourceType)
	}
	g.sourceMap[column.Source] = so
	return source.NewIndexer(so, column), nil
}

func (g *AIRowsGenerator) generate(ctx context.Context, batch int) ([]map[string]*schema.CellValue, error) {
	err := g.newBatch(ctx, batch)
	if err != nil {
		return nil, err
	}
	err = g.prepareRows(ctx, batch)
	if err != nil {
		return nil, err
	}
	// no more rows to be autofill
	if g.autofill.Enable && len(g.rows) == 0 {
		return nil, nil
	}
	chatRows := []map[string]any{}
	// used in autofill mode only
	contextColumnIDs := map[string]bool{}
	for _, col := range g.contextColumns {
		contextColumnIDs[col.Nanoid] = true
	}
	for _, row := range g.rows {
		cr := map[string]any{}
		for k, v := range row {
			if g.autofill.Enable && k != "id" {
				if _, ok := contextColumnIDs[k]; !ok {
					continue
				}
			}
			if v.ContextValue != nil {
				cr[k] = v.ContextValue
			} else {
				cr[k] = v.Value
			}
		}
		chatRows = append(chatRows, cr)
	}
	g.builder.AddTableColumns(g.table.Edges.Columns, g.autofill.Enable)
	g.builder.AddMissingColumns(g.missingColumns)
	err = g.builder.AddExistings(chatRows)
	if err != nil {
		return nil, err
	}
	generated := []map[string]*schema.CellValue{}
	resp, err := g.chat(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := util.TryDecodeJsonArray[map[string]any](
		gjson.Get(resp.Content, "data").String(),
	)
	// log error only because we need those successfully generated rows
	if err != nil {
		g.logger.Errorw("TryDecodeJsonArray error", "errpr", err)
	}

	// If the table schema contains pick-type columns, their cell values are pre-generated and stored in g.rows,
	// ensuring that len(g.rows) > 0. In this case, we assign generated column values to the existing rows.
	// Otherwise, we use the generated rows as complete rows directly.
	if len(g.rows) > 0 {
		for i, row := range g.rows {
			if len(rows) > i {
				if cast.ToString(rows[i]["id"]) != cast.ToString(row["id"].Value) {
					return nil, errors.New("generated row id mismatch")
				}
				for k, v := range rows[i] {
					row[k] = &schema.CellValue{Value: v}
				}
			}
			// on create rows, id is an internal columns to hint LLM to return correct count/order
			// on autofill rows, id is the database nanoid field, so must keep it so we can update database based on this id
			if !g.autofill.Enable {
				delete(row, "id")
			}
			generated = append(generated, row)
		}
	} else {
		for _, row := range rows {
			n := map[string]*schema.CellValue{}
			for k, v := range row {
				n[k] = &schema.CellValue{Value: v}
			}
			delete(row, "id")
			generated = append(generated, n)
		}
	}
	return generated, nil
}

func (g *AIRowsGenerator) Next(ctx context.Context) ([]map[string]*schema.CellValue, error) {
	if g.current >= g.total {
		return []map[string]*schema.CellValue{}, nil
	}
	batchSize := g.batchSize
	if left := g.total - g.current; batchSize > left {
		batchSize = left
	}
	rows, err := g.generate(ctx, batchSize)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return []map[string]*schema.CellValue{}, nil
	}
	if len(rows) > batchSize {
		rows = rows[:batchSize]
	}
	g.generated = append(g.generated, rows...)
	g.current += len(rows)
	if g.saveTo == "" {
		if g.autofill.Enable {
			// update rows by nanoid
			indexer := util.NewColumnIndexer(g.table.Edges.Columns)
			creates := []*ent.TableRowCreate{}
			for _, row := range rows {
				v, err := indexer.RowMapToSlice(row)
				if err != nil {
					return nil, err
				}
				creates = append(creates, g.db.TableRow.Create().SetNanoid(
					cast.ToString(row["id"].Value),
				).SetCells(v).SetTablemeta(g.table))
			}
			err = g.db.TableRow.CreateBulk(creates...).OnConflictColumns(tablerow.FieldNanoid).UpdateNewValues().Exec(ctx)
			if err != nil {
				return nil, err
			}
		} else {
			// create rows
			indexer := util.NewColumnIndexer(g.table.Edges.Columns)
			creates := []*ent.TableRowCreate{}
			for _, row := range rows {
				v, err := indexer.RowMapToSlice(row)
				if err != nil {
					return nil, err
				}
				creates = append(creates, g.db.TableRow.Create().SetCells(v).SetTablemeta(g.table))
			}
			err = g.db.TableRow.CreateBulk(creates...).Exec(ctx)
			if err != nil {
				return nil, err
			}
		}
	}
	return rows, nil
}

func (g *AIRowsGenerator) Table() *ent.TableMeta {
	return g.table
}
