package api

import (
	"tablepilot/config"
	"tablepilot/ent"
	"tablepilot/services/ai"
	"tablepilot/services/table"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type HTTPServer struct {
	config       *config.Config
	db           *ent.Client
	Logger       *zap.SugaredLogger
	aiService    ai.AiService
	tableService table.TableService
	apiv1        *gin.RouterGroup
}

func NewHttpServer(
	config *config.Config, db *ent.Client, debug bool,
	logger *zap.SugaredLogger, aiService ai.AiService, tableService table.TableService,
) *HTTPServer {
	if debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	return &HTTPServer{
		config:       config,
		db:           db,
		Logger:       logger,
		aiService:    aiService,
		tableService: tableService,
	}
}

func (hs *HTTPServer) RegisterRoutes(engine *gin.Engine) {
	engine.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowCredentials: true,
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
	}))
	hs.apiv1 = engine.Group("/v1")

	hs.addRouters()
}
