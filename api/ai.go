package api

import (
	"fmt"
	"net/http"

	"github.com/Yiling-J/tablepilot/services/ai"
	"github.com/gin-gonic/gin"
)

// AI-related handlers

func (hs *HTTPServer) GenerateListOptions(ctx *gin.Context) {
	var req ai.GenerateListOptionsRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		errorResponse(ctx, 400, err)
		return
	}

	options, err := hs.AIService.GenerateListOptions(ctx.Request.Context(), req)
	if err != nil {
		errorResponse(ctx, http.StatusInternalServerError, fmt.Errorf("failed to generate options %w", err))
		return
	}
	ctx.JSON(http.StatusOK, options)
}
