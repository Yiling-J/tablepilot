package api

import (
	"github.com/Yiling-J/tablepilot/services/table"
	"github.com/Yiling-J/tablepilot/services/table/util"

	"github.com/gin-gonic/gin"
)

func errorResponse(ctx *gin.Context, code int, err error) {
	_ = ctx.Error(err)
	ctx.JSON(code, err.Error())
}

func (hs *HTTPServer) CreateTable(ctx *gin.Context) {
	var request table.TableGenRequest
	err := ctx.ShouldBindJSON(&request)
	if err != nil {
		errorResponse(ctx, 400, err)
		return
	}

	uid, err := hs.TableService.CreateTable(ctx.Request.Context(), &request)
	if err != nil {
		errorResponse(ctx, 500, err)
		return
	}

	ctx.JSON(200, gin.H{"id": uid})
}

func (hs *HTTPServer) Generate(ctx *gin.Context) {
	var request table.GenerateRowsRequest
	err := ctx.ShouldBindJSON(&request)
	if err != nil {
		errorResponse(ctx, 400, err)
		return
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
		}
		if len(rows) == 0 {
			break
		}
		for _, row := range rows {
			dr, err := indexer.ToDisplayRow(row)
			if err != nil {
				errorResponse(ctx, 500, err)
				return
			}
			data = append(data, dr)
		}
	}
	ctx.JSON(200, gin.H{"data": data})
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
			r[col.Name] = row.Cells[i].Value
		}
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
	columns, err := hs.TableService.GetTableColumns(ctx.Request.Context(), ctx.Param("table"))
	if err != nil {
		errorResponse(ctx, 500, err)
		return
	}
	ctx.JSON(200, gin.H{"columns": columns})
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

func (hs *HTTPServer) addRouters() {
	hs.apiv1.POST("/tables", hs.CreateTable)
	hs.apiv1.GET("/tables", hs.ListTables)
	hs.apiv1.GET("/tables/:table", hs.Describe)
	hs.apiv1.DELETE("/tables/:table", hs.Delete)
	hs.apiv1.POST("/tables/:table/truncate", hs.Truncate)
	hs.apiv1.POST("/generate/tables/:table", hs.Generate)
	hs.apiv1.GET("/tables/:table/rows", hs.Rows)
}
