package api

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/Yiling-J/tablepilot/ent"
	"github.com/Yiling-J/tablepilot/ent/dataset"
	services_dataset "github.com/Yiling-J/tablepilot/services/dataset"
	"github.com/gin-gonic/gin"
)

// Dataset-related handlers

func (hs *HTTPServer) CreateDataset(ctx *gin.Context) {
	var apiReq services_dataset.DatasetAPIRequest
	if err := ctx.ShouldBind(&apiReq); err != nil {
		errorResponse(ctx, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	serviceReq := &services_dataset.CreateDatasetRequest{
		Name:        apiReq.Name,
		Description: apiReq.Description,
		Type:        dataset.Type(apiReq.Type),
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

	if err := ctx.ShouldBind(&apiReq); err != nil {
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
