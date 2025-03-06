package api

import (
	"time"

	"github.com/Yiling-J/tablepilot/services"

	"github.com/gin-contrib/cors"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
)

type HTTPServer struct {
	*services.Backend
	Engine *gin.Engine
	apiv1  *gin.RouterGroup
}

func NewHttpServer(backend *services.Backend, verbose bool) *HTTPServer {
	engine := gin.Default()
	if verbose {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	return &HTTPServer{
		Engine:  engine,
		Backend: backend,
	}
}

func (hs *HTTPServer) RegisterRoutes() {
	hs.Engine.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowCredentials: true,
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
	}))
	hs.Engine.Use(ginzap.Ginzap(hs.Logger.Desugar(), time.RFC3339, true))
	hs.Engine.Use(ginzap.RecoveryWithZap(hs.Logger.Desugar(), true))
	hs.apiv1 = hs.Engine.Group("/api/v1")
	hs.addRouters()
}
