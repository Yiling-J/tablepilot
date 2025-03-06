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

	ctx.JSON(200, map[string]string{"id": uid})
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
	ctx.JSON(200, map[string]any{"data": data})
}

func (hs *HTTPServer) addRouters() {
	hs.apiv1.POST("/tables", hs.CreateTable)
	hs.apiv1.POST("/generate/tables/:table", hs.Generate)
}
