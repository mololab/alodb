package web

import (
	agentApp "github.com/mololab/alodb/internal/application/agent"
	domainAgent "github.com/mololab/alodb/internal/domain/agent"
	"github.com/mololab/alodb/internal/infrastructure/config"
	"github.com/mololab/alodb/internal/infrastructure/web/handlers"

	"github.com/gin-gonic/gin"
)

type Server struct {
	router       *gin.Engine
	config       *config.Config
	agentService *agentApp.Service
}

func CORSMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Header("Access-Control-Allow-Origin", "*")
		ctx.Header("Access-Control-Allow-Credentials", "true")
		ctx.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-API-Key")
		ctx.Header("Access-Control-Allow-Methods", "POST, HEAD, PATCH, OPTIONS, GET, PUT, DELETE")

		if ctx.Request.Method == "OPTIONS" {
			ctx.AbortWithStatus(204)
			return
		}

		ctx.Next()
	}
}

func NewServer(cfg *config.Config) *Server {
	router := gin.Default()

	router.Use(CORSMiddleware())

	agentService := agentApp.NewService(domainAgent.AgentConfig{
		Providers:      cfg.Providers,
		SchemaCacheTTL: cfg.Agent.SchemaCacheTTL,
	})

	setupRoutes(router, agentService)

	return &Server{
		router:       router,
		config:       cfg,
		agentService: agentService,
	}
}

func setupRoutes(router *gin.Engine, agentService *agentApp.Service) {
	agentHandler := handlers.NewAgentHandler(agentService)

	v1 := router.Group("/v1")
	{
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "healthy"})
		})

		v1.GET("/models", agentHandler.GetModels)

		agent := v1.Group("/agent")
		{
			agent.POST("/chat", agentHandler.Chat)
		}
	}
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
