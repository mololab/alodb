package web

import (
	agentApp "github.com/mololab/alodb/internal/application/agent"
	domainAgent "github.com/mololab/alodb/internal/domain/agent"
	"github.com/mololab/alodb/internal/infrastructure/config"
	"github.com/mololab/alodb/internal/infrastructure/web/handlers"
	infraWS "github.com/mololab/alodb/internal/infrastructure/websocket"
	"github.com/mololab/alodb/pkg/logger"

	"github.com/gin-gonic/gin"
)

type Server struct {
	router       *gin.Engine
	config       *config.Config
	agentService *agentApp.Service
	wsHub        *infraWS.Hub
}

func CORSMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Header("Access-Control-Allow-Origin", "*")
		ctx.Header("Access-Control-Allow-Credentials", "true")
		ctx.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Gemini-Api-Key")
		ctx.Header("Access-Control-Allow-Methods", "POST, HEAD, PATCH, OPTIONS, GET, PUT, DELETE")

		if ctx.Request.Method == "OPTIONS" {
			ctx.AbortWithStatus(204)
			return
		}

		ctx.Next()
	}
}

// TODO: implement cron job to cleanup unused agents from manager cache based on lastUsed

func NewServer(cfg *config.Config) *Server {
	router := gin.Default()
	router.Use(CORSMiddleware())

	agentService := agentApp.NewService(domainAgent.AgentConfig{
		SchemaCacheTTL: cfg.Agent.SchemaCacheTTL,
	})

	wsHub := infraWS.NewHub()
	go wsHub.Run()

	setupRoutes(router, agentService, wsHub)

	return &Server{
		router:       router,
		config:       cfg,
		agentService: agentService,
		wsHub:        wsHub,
	}
}

func setupRoutes(router *gin.Engine, agentService *agentApp.Service, wsHub *infraWS.Hub) {
	agentHandler := handlers.NewAgentHandler(agentService)
	wsHandler := infraWS.NewHandler(wsHub, agentService.GetManager())

	v1 := router.Group("/v1")
	{
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "healthy"})
		})

		v1.GET("/models", agentHandler.GetModels)

		v1.GET("/agent/stream", func(c *gin.Context) {
			wsHandler.ServeWS(c.Writer, c.Request)
		})

		v1.DELETE("/sessions/:session_id", agentHandler.DeleteSession)
	}

	logger.Info().Msg("routes configured")
}

func (s *Server) Start() error {
	return s.router.Run(":" + s.config.Server.Port)
}

func (s *Server) Stop() error {
	if s.agentService != nil {
		return s.agentService.Close()
	}
	return nil
}
