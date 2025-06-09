package api

import (
	"errors"

	"github.com/Yiling-J/tablepilot/ent/schema"
	"github.com/Yiling-J/tablepilot/services/table/util" // For NewColumnIndexer
	"github.com/Yiling-J/tablepilot/services/workflow"
	"github.com/gin-gonic/gin"
)

// DecodeDataURL is assumed to be available from api/utils.go within package api.
// sseHeaders (if needed by RunWorkflow) will be called from table.go or a shared location.

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
				// Assuming DecodeDataURL is available in the package or globally
				content, err := DecodeDataURL(data) // This function needs to be defined/imported
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
			errorResponse(ctx, 500, err) // This might write a JSON error after SSE
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
					errorResponse(ctx, 500, err) // This might write a JSON error after SSE
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
						errorResponse(ctx, 500, err) // This might write a JSON error after SSE
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
	ctx.JSON(200, "") // This line was present in original, and test expects it.
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
