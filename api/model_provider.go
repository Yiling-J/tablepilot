package api

import (
	"github.com/Yiling-J/tablepilot/services/provider"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// Model and Provider related handlers

func (hs *HTTPServer) ListModels(ctx *gin.Context) {
	modelList := hs.AIService.ListModels(ctx.Request.Context())
	ctx.JSON(200, modelList)
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
	var p provider.Provider // Renamed from 'provider' to 'p' to avoid conflict with package name
	err := ctx.ShouldBindJSON(&p)
	if err != nil {
		errorResponse(ctx, 400, err)
		return
	}
	err = hs.ProviderService.CreateProvider(ctx.Request.Context(), p)
	if err != nil {
		errorResponse(ctx, 500, err)
		return
	}
	ctx.JSON(200, "")
}

func (hs *HTTPServer) UpdateProvider(ctx *gin.Context) {
	var p provider.Provider // Renamed from 'provider' to 'p' to avoid conflict with package name
	err := ctx.ShouldBindJSON(&p)
	if err != nil {
		errorResponse(ctx, 400, err)
		return
	}
	err = hs.ProviderService.UpdateProvider(ctx.Request.Context(), cast.ToInt(ctx.Param("id")), p)
	if err != nil {
		errorResponse(ctx, 500, err)
		return
	}
	ctx.JSON(200, "")
}
