package api

import (
	"github.com/Yiling-J/tablepilot/services/table"
	"github.com/Yiling-J/tablepilot/services/table/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Table-related handlers

func (hs *HTTPServer) CreateTable(ctx *gin.Context) {
	var request table.TableGenRequest
	err := ctx.ShouldBindJSON(&request)
	if err != nil {
		errorResponse(ctx, 400, err)
		return
	}
	request.MarkAPIRequest()

	uid, err := hs.TableService.Create(ctx.Request.Context(), &request)
	if err != nil {
		errorResponse(ctx, 500, err)
		return
	}

	ctx.JSON(200, gin.H{"id": uid})
}

func (hs *HTTPServer) UpdateTable(ctx *gin.Context) {
	var request table.TableGenRequest
	err := ctx.ShouldBindJSON(&request)
	if err != nil {
		errorResponse(ctx, 400, err)
		return
	}
	request.MarkAPIRequest()

	uid, err := hs.TableService.Update(ctx.Request.Context(), ctx.Param("table"), &request)
	if err != nil {
		errorResponse(ctx, 500, err)
		return
	}

	ctx.JSON(200, gin.H{"id": uid})
}

func sseHeaders(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")
}

func (hs *HTTPServer) gen(ctx *gin.Context, request table.GenerateRowsRequest) {
	if request.Stream {
		sseHeaders(ctx)
	}
	request.Table = ctx.Param("table")
	generator, err := hs.TableService.Genetate(ctx.Request.Context(), request)
	if err != nil {
		errorResponse(ctx, 500, err)
		return
	}
	indexer := util.NewColumnIndexer(generator.Table().Edges.Columns)
	data := []map[string]any{}
	for i := 0; ; i++ {
		hs.Logger.Debugw("start generating rows", "batch", i)
		rows, err := generator.Next(ctx.Request.Context())
		if err != nil {
			errorResponse(ctx, 500, err)
			return
		}
		if len(rows) == 0 {
			break
		}
		for _, row := range rows {
			dr, err := indexer.ToAPIRow(row)
			if err != nil {
				errorResponse(ctx, 500, err)
				return
			}

			data = append(data, dr)
		}
		if request.Stream {
			ctx.SSEvent("message", map[string]any{
				"data": data,
			})
			data = data[:0]
			ctx.Writer.Flush()
		}
	}
	if request.Stream {
		ctx.SSEvent("message", "[DONE]")
		ctx.Writer.Flush()
	}
	ctx.JSON(200, gin.H{"data": data})
}

func (hs *HTTPServer) Generate(ctx *gin.Context) {
	var request table.GenerateRowsRequest
	err := ctx.ShouldBindJSON(&request)
	if err != nil {
		errorResponse(ctx, 400, err)
		return
	}
	request.Autofill.Enable = false
	hs.gen(ctx, request)
}

func (hs *HTTPServer) Autofill(ctx *gin.Context) {
	var request table.GenerateRowsRequest
	err := ctx.ShouldBindJSON(&request)
	if err != nil {
		errorResponse(ctx, 400, err)
		return
	}
	request.Autofill.Enable = true
	if len(request.Autofill.ContextColumns) == 0 {
		// add a random not exist id
		request.Autofill.ContextColumns = []string{uuid.New().String()}
	}
	hs.gen(ctx, request)
}

func (hs *HTTPServer) Regenerate(ctx *gin.Context) {
	var request table.GenerateRowsRequest
	err := ctx.ShouldBindJSON(&request)
	if err != nil {
		errorResponse(ctx, 400, err)
		return
	}
	request.Autofill.Enable = true
	request.Count = len(request.Autofill.Rows)

	// add all columns as context columns
	tableObj, err := hs.TableService.GetTableDetail(ctx, ctx.Param("table")) // Renamed 'table' to 'tableObj' to avoid conflict
	if err != nil {
		errorResponse(ctx, 500, err)
		return
	}
	for _, col := range tableObj.Columns {
		request.Autofill.ContextColumns = append(request.Autofill.ContextColumns, col.ID)
	}
	hs.gen(ctx, request)
}

func (hs *HTTPServer) Rows(ctx *gin.Context) {
	rows, err := hs.TableService.Rows(ctx.Request.Context(), ctx.Param("table"))
	if err != nil {
		errorResponse(ctx, 500, err)
		return
	}
	data := []map[string]any{}
	for _, row := range rows.Rows {
		r := map[string]any{}
		for i, col := range rows.Columns {
			r[col.Nanoid] = row.Cells[i].Value
		}
		r["__id__"] = row.Nanoid
		data = append(data, r)
	}
	ctx.JSON(200, gin.H{"data": data, "total": len(data)})
}

func (hs *HTTPServer) ListTables(ctx *gin.Context) {
	tables, err := hs.TableService.ListTables(ctx.Request.Context())
	if err != nil {
		errorResponse(ctx, 500, err)
		return
	}
	ctx.JSON(200, tables)
}

func (hs *HTTPServer) Describe(ctx *gin.Context) {
	tableObj, err := hs.TableService.GetTableDetail(ctx.Request.Context(), ctx.Param("table")) // Renamed 'table' to 'tableObj'
	if err != nil {
		errorResponse(ctx, 500, err)
		return
	}
	ctx.JSON(200, tableObj)
}

func (hs *HTTPServer) Delete(ctx *gin.Context) {
	_, err := hs.TableService.Delete(ctx.Request.Context(), ctx.Param("table"))
	if err != nil {
		errorResponse(ctx, 500, err)
		return
	}
	ctx.JSON(204, nil)
}

func (hs *HTTPServer) Truncate(ctx *gin.Context) {
	n, err := hs.TableService.Truncate(ctx.Request.Context(), ctx.Param("table"))
	if err != nil {
		errorResponse(ctx, 500, err)
		return
	}
	ctx.JSON(200, gin.H{"removed": n})
}

func (hs *HTTPServer) CreateRows(ctx *gin.Context) {
	var request table.CreateRowsRequest
	err := ctx.ShouldBindJSON(&request)
	if err != nil {
		errorResponse(ctx, 400, err)
		return
	}
	err = hs.TableService.CreateRows(ctx.Request.Context(), ctx.Param("table"), request.Rows)
	if err != nil {
		errorResponse(ctx, 500, err)
		return
	}
	ctx.JSON(200, "")
}

func (hs *HTTPServer) SharedSources(ctx *gin.Context) {
	sources := hs.TableService.SharedSources(ctx)
	ctx.JSON(200, gin.H{"sources": sources})
}

func (hs *HTTPServer) GetTableSchema(ctx *gin.Context) {
	schema, err := hs.TableService.GetTableSchema(ctx.Request.Context(), ctx.Param("table"))
	if err != nil {
		errorResponse(ctx, 500, err)
		return
	}
	ctx.JSON(200, schema)
}

func (hs *HTTPServer) ImportImage(ctx *gin.Context) {
	var req table.ImportRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		errorResponse(ctx, 400, err)
		return
	}
	id, err := hs.TableService.ImportImage(ctx.Request.Context(), req)
	if err != nil {
		errorResponse(ctx, 500, err)
		return
	}
	ctx.JSON(200, gin.H{"id": id})
}
