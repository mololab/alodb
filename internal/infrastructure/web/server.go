package web

import (
	"context"

	agentApp "github.com/mololab/alodb/internal/application/agent"
	domainAgent "github.com/mololab/alodb/internal/domain/agent"
	"github.com/mololab/alodb/internal/infrastructure/config"
	"github.com/mololab/alodb/internal/infrastructure/web/handlers"

	"github.com/gin-gonic/gin"
)

// Server represents the HTTP server
type Server struct {
	router       *gin.Engine
	config       *config.Config
	agentService *agentApp.Service
}

// CORSMiddleware allows all origins for the initial state
// FUTURE TODO: make origin specific
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

// NewServer creates a new HTTP server
func NewServer(cfg *config.Config) *Server {
	router := gin.Default()

	router.Use(CORSMiddleware())

	// create agent service
	agentService := agentApp.NewService(domainAgent.AgentConfig{
		GoogleAPIKey:   cfg.Google.APIKey,
		SchemaCacheTTL: cfg.Agent.SchemaCacheTTL,
	})

	// setup routes
	setupRoutes(router, agentService)

	return &Server{
		router:       router,
		config:       cfg,
		agentService: agentService,
	}
}

// setupRoutes configures all the routes
func setupRoutes(router *gin.Engine, agentService *agentApp.Service) {
	agentHandler := handlers.NewAgentHandler(agentService)

	v1 := router.Group("/v1")
	{
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "healthy"})
		})

		agent := v1.Group("/agent")
		{
			agent.POST("/chat", agentHandler.Chat)
		}
	}
}

// Start starts the HTTP server
func (s *Server) Start(ctx context.Context) error {
	if err := s.agentService.Initialize(ctx); err != nil {
		return err
	}

	return s.router.Run(":" + s.config.Server.Port)
}

// Stop gracefully stops the server
func (s *Server) Stop() error {
	if s.agentService != nil {
		return s.agentService.Close()
	}
	return nil
}
