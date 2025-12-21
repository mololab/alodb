package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mololab/alodb/internal/infrastructure/config"
	"github.com/mololab/alodb/internal/infrastructure/web"
)

func main() {
	// load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// init web server
	server := web.NewServer(&cfg)
	log.Printf("Server starting on port %s", cfg.Server.Port)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := server.Start(ctx); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
		}
	}()

	// wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	cancel()

	if err := server.Stop(); err != nil {
		log.Printf("Error stopping server: %v", err)
	}

	time.Sleep(1 * time.Second)

	log.Println("Server exited")
}
