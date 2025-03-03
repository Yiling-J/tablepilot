package api

import (
	"github.com/Yiling-J/tablepilot/services/table"

	"github.com/gin-gonic/gin"
)

func (hs *HTTPServer) CreateTable(ctx *gin.Context) {
	var request table.TableGenRequest
	err := ctx.ShouldBindJSON(&request)
	if err != nil {
		ctx.JSON(400, err.Error())
		return
	}

	uid, err := hs.tableService.CreateTable(ctx.Request.Context(), &request)
	if err != nil {
		ctx.JSON(500, err.Error())
		return
	}

	ctx.JSON(200, map[string]string{"uid": uid})
}

func (hs *HTTPServer) addRouters() {
	hs.apiv1.POST("/tables", hs.CreateTable)
}
