package api

import (
	"github.com/Yiling-J/tablepilot/services/table"
	"github.com/Yiling-J/tablepilot/services/table/util"

	"github.com/gin-gonic/gin"
)

func (hs *HTTPServer) CreateTable(ctx *gin.Context) {
	var request table.TableGenRequest
	err := ctx.ShouldBindJSON(&request)
	if err != nil {
		ctx.JSON(400, err.Error())
		return
	}

	uid, err := hs.TableService.CreateTable(ctx.Request.Context(), &request)
	if err != nil {
		ctx.JSON(500, err.Error())
		return
	}

	ctx.JSON(200, map[string]string{"id": uid})
}

func (hs *HTTPServer) Generate(ctx *gin.Context) {
	var request table.GenerateRowsRequest
	err := ctx.ShouldBindJSON(&request)
	if err != nil {
		ctx.JSON(400, err.Error())
		return
	}
	request.Table = ctx.Param("table")
	generator, err := hs.TableService.Genetate(ctx.Request.Context(), request)
	if err != nil {
		ctx.JSON(500, err.Error())
		return
	}
	indexer := util.NewColumnIndexer(generator.Table().Edges.Columns)
	data := []map[string]any{}
	for {
		rows, err := generator.Next(ctx.Request.Context())
		if err != nil {
			ctx.JSON(500, err.Error())
		}
		if len(rows) == 0 {
			break
		}
		for _, row := range rows {
			dr, err := indexer.ToDisplayRow(row)
			if err != nil {
				ctx.JSON(500, err.Error())
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
