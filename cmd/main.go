package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mololab/alodb/internal/infrastructure/config"
	"github.com/mololab/alodb/internal/infrastructure/web"
	"github.com/mololab/alodb/pkg/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic("Failed to load configuration: " + err.Error())
	}

	logger.Init(cfg.Server.Env == "development")

	server := web.NewServer(&cfg)
	logger.Info().Str("port", cfg.Server.Port).Msg("server starting")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := server.Start(ctx); err != nil && err != http.ErrServerClosed {
			logger.Error().Err(err).Msg("server error")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info().Msg("shutting down server")

	cancel()

	if err := server.Stop(); err != nil {
		logger.Error().Err(err).Msg("error stopping server")
	}

	time.Sleep(1 * time.Second)

	logger.Info().Msg("server exited")
}
