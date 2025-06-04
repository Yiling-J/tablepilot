package api

import (
	"github.com/gin-gonic/gin"
)

func errorResponse(ctx *gin.Context, code int, err error) {
	_ = ctx.Error(err)
	ctx.JSON(code, err.Error())
}

func (hs *HTTPServer) addRouters() {
	hs.apiv1.GET("/models", hs.ListModels)
	hs.apiv1.POST("/tables", hs.CreateTable)
	hs.apiv1.GET("/tables", hs.ListTables)
	hs.apiv1.GET("/tables/:table", hs.Describe)
	hs.apiv1.PATCH("/tables/:table", hs.UpdateTable)
	hs.apiv1.POST("/tables/:table", hs.CreateRows)
	hs.apiv1.DELETE("/tables/:table", hs.Delete)
	hs.apiv1.POST("/tables/:table/truncate", hs.Truncate)
	hs.apiv1.POST("/generate/tables/:table", hs.Generate)
	hs.apiv1.POST("/autofill/tables/:table", hs.Autofill)
	hs.apiv1.GET("/tables/:table/rows", hs.Rows)
	hs.apiv1.GET("/sources", hs.SharedSources)
	hs.apiv1.GET("/tables/:table/schema", hs.GetTableSchema)
	hs.apiv1.GET("/providers", hs.GetProviders)
	hs.apiv1.POST("/providers", hs.CreateProvider)
	hs.apiv1.DELETE("/providers/:id", hs.DeleteProvider)
	hs.apiv1.PATCH("/providers/:id", hs.UpdateProvider)
	hs.apiv1.POST("/regenerate/tables/:table", hs.Regenerate)
	hs.apiv1.POST("/image_import/tables", hs.ImportImage)

	// ai
	hs.apiv1.POST("/ai/list_gen", hs.GenerateListOptions)

	// workflows
	hs.apiv1.GET("/workflows", hs.ListWorkflows)
	hs.apiv1.GET("/workflows/:id", hs.GetWorkflow)
	hs.apiv1.POST("/workflows/:workflow/run", hs.RunWorkflow)
	hs.apiv1.POST("/workflows", hs.CreateWorkflow)
	hs.apiv1.PATCH("/workflows/:id", hs.UpdateWorkflow)
	hs.apiv1.DELETE("/workflows/:id", hs.DeleteWorkflow)

	// dataset
	datasetRoutes := hs.apiv1.Group("/datasets")
	datasetRoutes.POST("", hs.CreateDataset)
	datasetRoutes.GET("", hs.ListDatasets)
	datasetRoutes.GET("/:id", hs.GetDataset)
	datasetRoutes.PATCH("/:id", hs.UpdateDataset)
	datasetRoutes.DELETE("/:id", hs.DeleteDataset)
	datasetRoutes.GET("/:id/preview", hs.PreviewDataset)
}
