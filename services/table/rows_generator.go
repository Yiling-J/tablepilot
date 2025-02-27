package table

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"tablepilot/ent"
	"tablepilot/ent/schema"
	"tablepilot/ent/tablecolumn"
	"tablepilot/ent/tablemeta"
	"tablepilot/ent/tablerow"
	"tablepilot/services/ai"
	"tablepilot/services/ai/client"
	"tablepilot/services/ai/promptbuilder"
	"tablepilot/services/table/source"
	"tablepilot/services/table/util"

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
	sourceMap      map[string]source.Source
	generated      []map[string]*schema.CellValue
	contextLength  int
	saveTo         string
	temperature    float64
	model          string

	total     int
	batchSize int
	current   int

	rows    []map[string]*schema.CellValue
	builder *promptbuilder.RowsBuilder
}

func NewRowsGenerator(ctx context.Context, params GenerateRowsParams, db *ent.Client, ai ai.AiService, logger *zap.SugaredLogger) (*AIRowsGenerator, error) {
	generator := &AIRowsGenerator{
		logger: logger,
		db:     db,
		ai:     ai,

		total:       params.Count,
		batchSize:   params.Batch,
		sourceMap:   make(map[string]source.Source),
		saveTo:      params.SaveTo,
		temperature: params.Temperature,
		model:       params.Model,
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

	for _, c := range meta.Edges.Columns {
		if c.ContextLength > generator.contextLength {
			generator.contextLength = c.ContextLength
		}
		if c.FillMode == tablecolumn.FillModePick {
			if len(c.Source) == 0 {
				return nil, errors.New("")
			}
			var so source.Source
			so, err := generator.columnSource(ctx, c)
			if err != nil {
				return nil, err
			}
			generator.sourceMap[c.Nanoid] = so
			continue
		}
		generator.missingColumns = append(generator.missingColumns, c)
	}
	return generator, nil
}

func (g *AIRowsGenerator) newBatch(ctx context.Context, batch int) error {
	g.builder = promptbuilder.NewRowsBuilder(batch)
	g.rows = g.rows[:0]
	return g.prepareContextRows(ctx)
}

func (g *AIRowsGenerator) prepareContextRows(ctx context.Context) error {
	// get required rows from previous generated results or database
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
	// add context values for each column
	for _, col := range g.table.Edges.Columns {
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

func (g *AIRowsGenerator) prepareRow(ctx context.Context) error {
	row := map[string]*schema.CellValue{}
	for _, col := range g.table.Edges.Columns {
		so, ok := g.sourceMap[col.Nanoid]
		if ok {
			v, err := so.Next(ctx)
			if err != nil {
				return err
			}
			row[col.Nanoid] = v
		}
	}
	if len(row) > 0 {
		g.rows = append(g.rows, row)
	}
	return nil
}

func (g *AIRowsGenerator) chat(ctx context.Context) (*client.ChatResponse, error) {
	g.builder.AddMissingColumns(g.missingColumns)
	om := orderedmap.New[string, *jsonschema.Schema]()
	om.Set("id", &jsonschema.Schema{Type: "integer"})
	required := []string{}
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

func (g *AIRowsGenerator) columnSource(ctx context.Context, column *ent.TableColumn) (source.Source, error) {
	var so source.Source
	sourceType := gjson.GetBytes(column.Source, "type").String()
	switch sourceType {
	case "list":
		var ls source.ListSource
		err := json.Unmarshal(column.Source, &ls)
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
		err := json.Unmarshal(column.Source, &ls)
		if err != nil {
			return nil, err
		}
		err = ls.Init(ctx, g.ai, column)
		if err != nil {
			return nil, err
		}
		so = &ls
	case "linked":
		var ls source.LinkedSource
		err := json.Unmarshal(column.Source, &ls)
		if err != nil {
			return nil, err
		}
		err = ls.Init(ctx, g.db, column)
		if err != nil {
			return nil, err
		}
		so = &ls
	default:
		return nil, fmt.Errorf("unknow source type %s", sourceType)
	}
	return so, nil
}

func (g *AIRowsGenerator) generate(ctx context.Context, batch int) ([]map[string]*schema.CellValue, error) {
	err := g.newBatch(ctx, batch)
	if err != nil {
		return nil, err
	}
	for n := 0; n < batch; n++ {
		err = g.prepareRow(ctx)
		if err != nil {
			return nil, err
		}
	}
	chatRows := []map[string]any{}
	for _, row := range g.rows {
		cr := map[string]any{}
		for k, v := range row {
			if v.ContextValue != nil {
				cr[k] = v.ContextValue
			} else {
				cr[k] = v.Value
			}
		}
		chatRows = append(chatRows, cr)
	}
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
	if err != nil && len(rows) == 0 {
		return nil, err
	}
	if err != nil {
		g.logger.Errorw("err happen when decode row gen response json", "err", err)
	}
	if len(g.rows) > 0 {
		for i, row := range g.rows {
			if len(rows) >= i {
				for k, v := range rows[i] {
					row[k] = &schema.CellValue{Value: v}
				}
			}
			generated = append(generated, row)
		}
	} else {
		for _, row := range rows {
			n := map[string]*schema.CellValue{}
			for k, v := range row {
				n[k] = &schema.CellValue{Value: v}
			}
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
	if len(rows) > batchSize {
		rows = rows[:batchSize]
	}
	g.generated = append(g.generated, rows...)
	g.current += len(rows)
	if g.saveTo == "" {
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
	return rows, nil
}

func (g *AIRowsGenerator) Table() *ent.TableMeta {
	return g.table
}
