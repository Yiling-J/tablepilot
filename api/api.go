package api

import (
	"errors"

	"fmt"
	"io"
	"net/http"

	"github.com/Yiling-J/tablepilot/ent/schema"
	"github.com/Yiling-J/tablepilot/services/provider"
	"github.com/Yiling-J/tablepilot/services/table"

	"github.com/Yiling-J/tablepilot/ent"
	services_dataset "github.com/Yiling-J/tablepilot/services/dataset"
	"github.com/Yiling-J/tablepilot/services/table/util"
	"github.com/Yiling-J/tablepilot/services/workflow"
	"github.com/google/uuid"
	"github.com/spf13/cast"

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
	table, err := hs.TableService.GetTableDetail(ctx, ctx.Param("table"))
	if err != nil {
		errorResponse(ctx, 500, err)
		return
	}
	for _, col := range table.Columns {
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
	table, err := hs.TableService.GetTableDetail(ctx.Request.Context(), ctx.Param("table"))
	if err != nil {
		errorResponse(ctx, 500, err)
		return
	}
	ctx.JSON(200, table)
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

func (hs *HTTPServer) ListModels(ctx *gin.Context) {
	modelList := hs.AIService.ListModels(ctx.Request.Context())
	ctx.JSON(200, modelList)
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

func (hs *HTTPServer) GetProviders(ctx *gin.Context) {
	providers, err := hs.ProviderService.ListProviders(ctx.Request.Context())
	if err != nil {
		errorResponse(ctx, 500, err)
		return
	}
	ctx.JSON(200, providers)
}

func (hs *HTTPServer) DeleteProvider(ctx *gin.Context) {
	err := hs.ProviderService.DeleteProvider(ctx.Request.Context(), cast.ToInt(ctx.Param("id")))
	if err != nil {
		errorResponse(ctx, 500, err)
		return
	}
	ctx.JSON(200, "")
}

func (hs *HTTPServer) CreateProvider(ctx *gin.Context) {
	var provider provider.Provider
	err := ctx.ShouldBindJSON(&provider)
	if err != nil {
		errorResponse(ctx, 400, err)
		return
	}
	err = hs.ProviderService.CreateProvider(ctx.Request.Context(), provider)
	if err != nil {
		errorResponse(ctx, 500, err)
		return
	}
	ctx.JSON(200, "")
}

func (hs *HTTPServer) UpdateProvider(ctx *gin.Context) {
	var provider provider.Provider
	err := ctx.ShouldBindJSON(&provider)
	if err != nil {
		errorResponse(ctx, 400, err)
		return
	}
	err = hs.ProviderService.UpdateProvider(ctx.Request.Context(), cast.ToInt(ctx.Param("id")), provider)
	if err != nil {
		errorResponse(ctx, 500, err)
		return
	}
	ctx.JSON(200, "")
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

func (hs *HTTPServer) ListWorkflows(ctx *gin.Context) {
	wfs, err := hs.WorkflowService.List(ctx.Request.Context())
	if err != nil {
		errorResponse(ctx, 500, err)
		return
	}
	r := []workflow.WorkflowSimple{}
	for _, w := range wfs {
		r = append(r, workflow.WorkflowSimple{
			ID:          w.Nanoid,
			Name:        w.Name,
			Description: w.Description,
		})
	}
	ctx.JSON(200, gin.H{"total": len(r), "workflows": r})
}

func (hs *HTTPServer) GetWorkflow(ctx *gin.Context) {
	wf, err := hs.WorkflowService.Get(ctx.Request.Context(), ctx.Param("id"))
	if err != nil {
		errorResponse(ctx, 500, err)
		return
	}
	ctx.JSON(200, workflow.Workflow{
		ID:          wf.Nanoid,
		Name:        wf.Name,
		Description: wf.Description,
		Variables:   wf.Variables,
		Steps:       wf.Steps,
	})
}

func (hs *HTTPServer) RunWorkflow(ctx *gin.Context) {
	var request workflow.StartWorklfowRequest
	err := ctx.ShouldBindJSON(&request)
	if err != nil {
		errorResponse(ctx, 400, err)
		return
	}
	// file will be encoded as data url, decode them first
	wf, err := hs.WorkflowService.Get(ctx.Request.Context(), ctx.Param("workflow"))
	if err != nil {
		errorResponse(ctx, 500, err)
		return
	}
	for _, va := range wf.Variables {
		if va.Type == schema.WorkflowVariableTypeFile {
			if v, ok := request.Variables[va.Name]; ok {
				vt, ok := v.(map[string]any)
				if !ok {
					errorResponse(ctx, 500, errors.New("invalid file input"))
					return
				}
				name, ok := vt["name"].(string)
				if !ok {
					errorResponse(ctx, 500, errors.New("invalid file input"))
					return
				}
				data, ok := vt["data"].(string)
				if !ok {
					errorResponse(ctx, 500, errors.New("invalid file input"))
					return
				}
				content, err := DecodeDataURL(data)
				if err != nil {
					errorResponse(ctx, 500, err)
					return
				}
				request.Variables[name+"__data"] = workflow.FileInfo{
					Data: content,
					Name: name,
				}
				request.Variables[va.Name] = name
			}
		}
	}
	sseHeaders(ctx)
	runner, err := hs.WorkflowService.Start(ctx, ctx.Param("workflow"), request)
	if err != nil {
		errorResponse(ctx, 500, err)
		return
	}
	for {
		result, err := runner.Next(ctx.Request.Context())
		if err != nil {
			ctx.SSEvent("message", map[string]any{
				"type": "ERROR",
				"data": err.Error(),
			})
			ctx.Writer.Flush()
			errorResponse(ctx, 500, err)
			return
		}
		if result == nil {
			break
		}
		switch result.Action {
		case workflow.WorkflowActionShowMessage:
			ctx.SSEvent("message", map[string]any{
				"type": "MESSAGE",
				"data": result.Message,
			})
			ctx.Writer.Flush()
		case workflow.WorkflowActionExport:
			ctx.SSEvent("message", map[string]any{
				"type": "EXPORT",
				"data": result.ExportData,
			})
			ctx.Writer.Flush()
			ctx.SSEvent("message", map[string]any{
				"type": "MESSAGE",
				"data": result.Message,
			})
			ctx.Writer.Flush()
		case workflow.WorkflowActionGenerate:
			ctx.SSEvent("message", map[string]any{
				"type": "MESSAGE",
				"data": result.Message,
			})
			ctx.Writer.Flush()
			generator := result.Generator
			indexer := util.NewColumnIndexer(generator.Table().Edges.Columns)
			for i := 0; ; i++ {
				hs.Logger.Debugw("start generating rows", "batch", i)
				rows, err := generator.Next(ctx.Request.Context())
				if err != nil {
					ctx.SSEvent("message", map[string]any{
						"type": "ERROR",
						"data": err.Error(),
					})
					ctx.Writer.Flush()
					errorResponse(ctx, 500, err)
					return
				}
				if len(rows) == 0 {
					break
				}
				data := []map[string]util.CellValueTyped{}
				for _, row := range rows {
					dr, err := indexer.ToAPIRowWIthType(row)
					if err != nil {
						ctx.SSEvent("message", map[string]any{
							"type": "ERROR",
							"data": err.Error(),
						})
						ctx.Writer.Flush()
						errorResponse(ctx, 500, err)
						return
					}

					data = append(data, dr)
				}
				ctx.SSEvent("message", map[string]any{
					"type": "ROWS",
					"data": data,
				})
				ctx.Writer.Flush()
			}
		}
		ctx.SSEvent("message", map[string]any{
			"type": "STEP_DONE",
		})
		ctx.Writer.Flush()
	}
	ctx.SSEvent("message", map[string]any{
		"type": "WORKFLOW_DONE",
	})
	ctx.Writer.Flush()
	ctx.SSEvent("message", "[DONE]")
	ctx.Writer.Flush()
	ctx.JSON(200, "")
}

func (hs *HTTPServer) CreateWorkflow(ctx *gin.Context) {
	var wf workflow.Workflow
	err := ctx.ShouldBindJSON(&wf)
	if err != nil {
		errorResponse(ctx, 400, err)
		return
	}
	id, err := hs.WorkflowService.Create(ctx.Request.Context(), &wf)
	if err != nil {
		errorResponse(ctx, 500, err)
		return
	}
	ctx.JSON(200, gin.H{"id": id})
}

func (hs *HTTPServer) UpdateWorkflow(ctx *gin.Context) {
	var wf workflow.Workflow
	err := ctx.ShouldBindJSON(&wf)
	if err != nil {
		errorResponse(ctx, 400, err)
		return
	}
	id, err := hs.WorkflowService.Update(ctx.Request.Context(), ctx.Param("id"), &wf)
	if err != nil {
		errorResponse(ctx, 500, err)
		return
	}
	ctx.JSON(200, gin.H{"id": id})
}

func (hs *HTTPServer) DeleteWorkflow(ctx *gin.Context) {
	err := hs.WorkflowService.Delete(ctx.Request.Context(), ctx.Param("id"))
	if err != nil {
		errorResponse(ctx, 500, err)
		return
	}
	ctx.JSON(200, "")
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

func (hs *HTTPServer) CreateDataset(ctx *gin.Context) {
	var apiReq services_dataset.DatasetAPIRequest
	if err := ctx.ShouldBind(&apiReq); err != nil {
		errorResponse(ctx, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	serviceReq := &services_dataset.CreateDatasetRequest{
		Name:        apiReq.Name,
		Description: apiReq.Description,
		Type:        apiReq.Type,
		Data:        apiReq.Data,
	}

	if apiReq.Type == "csv" {
		if len(apiReq.Files) == 0 {
			errorResponse(ctx, http.StatusBadRequest, errors.New("at least one file is required for CSV dataset type"))
			return
		}
		var readers []io.Reader
		for _, fh := range apiReq.Files {
			f, err := fh.Open()
			if err != nil {
				errorResponse(ctx, http.StatusBadRequest, err)
				return
			}
			readers = append(readers, f)
		}
		serviceReq.Files = readers
	}

	nanoid, err := hs.DatasetService.Create(ctx.Request.Context(), serviceReq)
	if err != nil {
		// Consider checking for specific error types, e.g., validation errors vs. server errors
		errorResponse(ctx, http.StatusInternalServerError, fmt.Errorf("failed to create dataset: %w", err))
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"id": nanoid, "name": serviceReq.Name})
}

func (hs *HTTPServer) GetDataset(ctx *gin.Context) {
	datasetID := ctx.Param("id")
	dsInfo, err := hs.DatasetService.Get(ctx.Request.Context(), datasetID)
	if err != nil {
		if ent.IsNotFound(err) {
			errorResponse(ctx, http.StatusNotFound, fmt.Errorf("dataset '%s' not found: %w", datasetID, err))
		} else {
			errorResponse(ctx, http.StatusInternalServerError, fmt.Errorf("failed to get dataset '%s': %w", datasetID, err))
		}
		return
	}
	ctx.JSON(http.StatusOK, dsInfo)
}

func (hs *HTTPServer) ListDatasets(ctx *gin.Context) {
	datasets, err := hs.DatasetService.List(ctx.Request.Context())
	if err != nil {
		errorResponse(ctx, http.StatusInternalServerError, fmt.Errorf("failed to list datasets: %w", err))
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"total": len(datasets), "datasets": datasets})
}

func (hs *HTTPServer) UpdateDataset(ctx *gin.Context) {
	datasetID := ctx.Param("id")
	var apiReq services_dataset.DatasetAPIRequest

	if err := ctx.ShouldBindJSON(&apiReq); err != nil {
		errorResponse(ctx, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	fields := []string{}
	if apiReq.Name != "" {
		fields = append(fields, "name")
	}
	if apiReq.Description != "" {
		fields = append(fields, "description")
	}
	if len(apiReq.Data) > 0 {
		fields = append(fields, "data")
	}

	serviceReq := &services_dataset.UpdateDatasetRequest{
		CreateDatasetRequest: services_dataset.CreateDatasetRequest{
			Name:        apiReq.Name,
			Description: apiReq.Description,
			Data:        apiReq.Data,
		},
		Fields: fields,
	}

	if apiReq.Files != nil {
		var readers []io.Reader
		for _, fh := range apiReq.Files {
			f, err := fh.Open()
			if err != nil {
				errorResponse(ctx, http.StatusBadRequest, err)
				return
			}
			readers = append(readers, f)
		}
		serviceReq.Files = readers
		serviceReq.Fields = append(serviceReq.Fields, "files")
	} else {
		serviceReq.Files = []io.Reader{}
	}

	err := hs.DatasetService.Update(ctx.Request.Context(), datasetID, serviceReq)
	if err != nil {
		if ent.IsNotFound(err) {
			errorResponse(ctx, http.StatusNotFound, fmt.Errorf("dataset '%s' not found for update: %w", datasetID, err))
		} else {
			errorResponse(ctx, http.StatusInternalServerError, fmt.Errorf("failed to update dataset '%s': %w", datasetID, err))
		}
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"id": datasetID})
}

func (hs *HTTPServer) DeleteDataset(ctx *gin.Context) {
	datasetID := ctx.Param("id")
	err := hs.DatasetService.Delete(ctx.Request.Context(), datasetID)
	if err != nil {
		if ent.IsNotFound(err) {
			errorResponse(ctx, http.StatusNotFound, fmt.Errorf("dataset '%s' not found for deletion: %w", datasetID, err))
		} else {
			errorResponse(ctx, http.StatusInternalServerError, fmt.Errorf("failed to delete dataset '%s': %w", datasetID, err))
		}
		return
	}
	ctx.JSON(http.StatusOK, "")
}

func (hs *HTTPServer) PreviewDataset(ctx *gin.Context) {
	datasetID := ctx.Param("id")
	previewData, err := hs.DatasetService.Preview(ctx.Request.Context(), datasetID)
	if err != nil {
		if ent.IsNotFound(err) {
			errorResponse(ctx, http.StatusNotFound, fmt.Errorf("dataset '%s' not found for preview: %w", datasetID, err))
		} else {
			errorResponse(ctx, http.StatusInternalServerError, fmt.Errorf("failed to preview dataset '%s': %w", datasetID, err))
		}
		return
	}
	ctx.JSON(http.StatusOK, previewData)
}
