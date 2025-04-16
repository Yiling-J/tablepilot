package table

import (
	"context"
	"path/filepath"
	"time"

	// #nosec
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"

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
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"go.uber.org/zap"
)

type RowsGenerator interface {
	Next(ctx context.Context) ([]map[string]*schema.CellValue, error)
	Table() *ent.TableMeta
}

type AIRowsGenerator struct {
	db                  *ent.Client
	ai                  ai.AiService
	logger              *zap.SugaredLogger
	table               *ent.TableMeta
	missingColumns      []*ent.TableColumn
	missingImageColumns []*ent.TableColumn
	contextColumns      []*ent.TableColumn
	indexerMap          map[string]*source.Indexer
	sourceMap           map[string]source.Source
	generated           []map[string]*schema.CellValue
	images              map[string]string
	contextLength       int
	saveTo              string
	temperature         float64
	model               string
	imageModel          string

	total     int
	batchSize int
	current   int
	offset    int

	rows    []map[string]*schema.CellValue
	builder *promptbuilder.RowsBuilder

	autofill      AutofillRequest
	sourceDataDir string
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
		imageModel:  params.ImageModel,
		autofill:    params.Autofill,
		offset:      params.Autofill.Offset,
		images:      make(map[string]string),
	}
	if params.sourceDataDir == "" {
		params.sourceDataDir = "./"
	}
	generator.sourceDataDir = params.sourceDataDir
	meta, err := db.TableMeta.Query().WithColumns(func(tcq *ent.TableColumnQuery) {
		tcq.Order(ent.Asc(tablecolumn.FieldID))
	}).Where(tablemeta.Or(
		tablemeta.Nanoid(params.Table),
		tablemeta.Name(params.Table),
	)).First(ctx)
	if err != nil {
		return nil, err
	}
	// add shared sources
	if meta.Sources == nil {
		meta.Sources = map[string]json.RawMessage{}
	}
	for name, source := range params.sharedSources {
		if _, ok := meta.Sources[name]; !ok {
			meta.Sources[name] = source
		}
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
		if c.Type != tablecolumn.TypeImage {
			generator.missingColumns = append(generator.missingColumns, c)
		} else {
			generator.missingImageColumns = append(generator.missingImageColumns, c)
		}
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

func (g *AIRowsGenerator) imageURL(ctx context.Context, raw string) (string, error) {
	_, err := url.ParseRequestURI(raw)
	if err == nil {
		return raw, nil
	}
	// already data url
	if strings.HasPrefix(raw, "data:") {
		return raw, nil
	}
	// try load image file
	root, err := os.OpenRoot(g.sourceDataDir)
	if err != nil {
		return "", err
	}
	f, err := root.Open(raw)
	if err != nil {
		return "", err
	}
	data, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}
	ct := http.DetectContentType(data)
	b64 := base64.StdEncoding.EncodeToString(data)
	if !strings.HasPrefix(ct, "image/") {
		return "", errors.New("not an image")
	}
	return fmt.Sprintf("data:%s;base64,%s", ct, b64), nil
}

func (g *AIRowsGenerator) prepareRows(ctx context.Context, batch int) error {
	g.images = map[string]string{}
	if g.autofill.Enable {
		rows, err := g.table.QueryRows().Order(
			ent.Asc(tablerow.FieldID),
		).Limit(batch).Offset(g.offset).All(ctx)
		if err != nil {
			return err
		}
		g.offset += len(rows)
		contextColumns := map[string]bool{}
		for _, col := range g.autofill.ContextColumns {
			contextColumns[col] = true
		}
		for _, dbrow := range rows {
			row := map[string]*schema.CellValue{}
			for i, col := range g.table.Edges.Columns {
				if col.Type == tablecolumn.TypeImage {
					v := cast.ToString(dbrow.Cells[i].Value)
					if v == "" {
						continue
					}
					// use md5 if v is too long(data url)
					vk := v
					if strings.HasPrefix(v, "data:") && len(v) > 200 {
						// #nosec
						vk = fmt.Sprintf("%x", md5.Sum([]byte(v)))
					}
					if _, ok := g.images[vk]; !ok {
						// don't actually read the image data, if the column is not in context columns
						if _, ok := contextColumns[col.Nanoid]; !ok {
							if _, ok := contextColumns[col.Name]; !ok {
								continue
							}
						}
						url, err := g.imageURL(ctx, v)
						if err != nil {
							return err
						}
						g.images[vk] = url
					}
					row[col.Nanoid] = &schema.CellValue{Value: vk}
				} else {
					row[col.Nanoid] = dbrow.Cells[i]
				}
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
				if col.Type == tablecolumn.TypeImage && ok {
					v := cast.ToString(row[col.Nanoid].Value)
					if v == "" {
						continue
					}
					// use md5 if v is too long(data url)
					vk := v
					if strings.HasPrefix(v, "data:") && len(v) > 200 {
						// #nosec
						vk = fmt.Sprintf("%x", md5.Sum([]byte(v)))
					}
					if _, ok := g.images[vk]; !ok {
						url, err := g.imageURL(ctx, v)
						if err != nil {
							return err
						}
						g.images[vk] = url
					}
					row[col.Nanoid] = &schema.CellValue{Value: vk}
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
	if len(g.images) > 0 {
		input = client.UserMessageWithImages(prompt, g.images)
	}
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

func (g *AIRowsGenerator) imageGen(ctx context.Context) (*client.ImageGenResponse, error) {
	prompt, err := g.builder.ImageGenPrompt()
	if err != nil {
		return nil, err
	}
	input := client.UserMessage(prompt)
	if len(g.images) > 0 {
		input = client.UserMessageWithImages(prompt, g.images)
	}
	temperature := g.temperature
	if temperature < 0 {
		temperature = GENERATE_DATA_TEMPERATURE
	}
	model := g.imageModel
	return g.ai.ImageGen(ctx, &client.ChatRequest{
		Temperature: temperature,
		Messages:    []*client.Message{input},
		Model:       model,
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
		err = ls.Init(ctx, g.sourceDataDir)
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
		err = ls.Init(ctx, g.logger, g.sourceDataDir)
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
				g.logger,
			)
		}
		err = ls.Init(ctx, client, g.logger, g.sourceDataDir)
		if err != nil {
			return nil, err
		}
		so = &ls
	case "files":
		var ls source.FilesSource
		err := json.Unmarshal(raw, &ls)
		if err != nil {
			return nil, err
		}
		err = ls.Init(ctx, g.logger, g.sourceDataDir)
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
	g.builder.AddMissingColumns(g.missingColumns, true)
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

func (g *AIRowsGenerator) generateImages(ctx context.Context, rows []map[string]*schema.CellValue) ([]map[string]*schema.CellValue, error) {
	err := g.newBatch(ctx, len(rows))
	if err != nil {
		return nil, err
	}
	chatRows := []map[string]any{}
	// used in autofill mode only
	contextColumnIDs := map[string]bool{}
	for _, col := range g.contextColumns {
		contextColumnIDs[col.Nanoid] = true
	}
	// In the autofill case, also include generated text columns as context.
	// Ideally, text and image are generated together and should provide mutual context.
	// For example, if you have a table with recipe names and steps, and you want to autofill
	// the ingredients and image, then the ingredients column should be considered part of the
	// context, even if it wasn't explicitly specified.
	if g.autofill.Enable {
		for _, col := range g.missingColumns {
			contextColumnIDs[col.Nanoid] = true
		}
	}
	idMap := map[string]int{}
	for i, row := range rows {
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
		var rowid string
		if v, ok := row["id"]; !ok {
			rowid = cast.ToString(i)
			row["id"] = &schema.CellValue{Value: i}
		} else {
			rowid = cast.ToString(v.Value)
		}
		idMap[rowid] = i
		cr["id"] = rowid
		chatRows = append(chatRows, cr)
	}
	g.builder.AddTableColumns(g.table.Edges.Columns, g.autofill.Enable)
	g.builder.AddMissingColumns(g.missingImageColumns, false)
	err = g.builder.AddExistings(chatRows)
	if err != nil {
		return nil, err
	}

	resp, err := g.imageGen(ctx)
	if err != nil {
		return nil, err
	}
	for id, data := range resp.Images {
		tmp := strings.Split(id, "-")
		row, column := tmp[0], tmp[1]
		imagesDir := filepath.Join(g.sourceDataDir, "tablepilot_images", g.table.Nanoid)
		if err := os.MkdirAll(imagesDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %q: %w", imagesDir, err)
		}
		if index, ok := idMap[row]; ok {
			root, err := os.OpenRoot(g.sourceDataDir)
			if err != nil {
				return nil, err
			}
			fp := fmt.Sprintf("tablepilot_images/%s/%s-%d.png", g.table.Nanoid, id, time.Now().Unix())
			f, err := root.Create(fp)
			if err != nil {
				return nil, err
			}
			defer func() { _ = f.Close() }()
			_, err = f.Write(data)
			if err != nil {
				return nil, err
			}
			row := rows[index]
			row[column] = &schema.CellValue{Value: fp}
		}
	}
	return rows, nil
}

func (g *AIRowsGenerator) Next(ctx context.Context) ([]map[string]*schema.CellValue, error) {
	if g.current >= g.total {
		return []map[string]*schema.CellValue{}, nil
	}
	batchSize := g.batchSize
	if left := g.total - g.current; batchSize > left {
		batchSize = left
	}
	rows := g.rows
	var err error
	if len(g.missingColumns) > 0 {
		rows, err = g.generate(ctx, batchSize)
		if err != nil {
			return nil, err
		}
		if len(rows) == 0 {
			return []map[string]*schema.CellValue{}, nil
		}
		if len(rows) > batchSize {
			rows = rows[:batchSize]
		}
	}
	if len(g.missingImageColumns) > 0 {
		// prepareRows step is done in generate function, but when missingColumns length is zero,
		// the generate method is skipped, so need to call prepareRows so rows is not empty
		if len(g.missingColumns) == 0 {
			err = g.prepareRows(ctx, batchSize)
			if err != nil {
				return nil, err
			}
			rows = g.rows
		}
		rows, err = g.generateImages(ctx, rows)
		if err != nil {
			return nil, err
		}
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
