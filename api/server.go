package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/Yiling-J/tablepilot/services"
	"github.com/Yiling-J/tablepilot/ui"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/static"
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

type indexHTML struct {
	fs http.FileSystem
}

func (e *indexHTML) Exists(prefix string, path string) bool {
	return !strings.HasPrefix(path, "/api/")
}

func (e *indexHTML) Open(path string) (http.File, error) {
	return e.fs.Open("dist/index.html")
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
	hs.Engine.Use(static.Serve("/", static.EmbedFolder(ui.Dist, "dist")))
	hs.Engine.Use(static.Serve("/", &indexHTML{fs: http.FS(ui.Dist)}))
	hs.apiv1 = hs.Engine.Group("/api/v1")
	hs.apiv1.Static("/images", hs.Config.Common.SourceDataDir)
	hs.addRouters()
}
