package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/raufhm/vfc/internal/config"
	"github.com/raufhm/vfc/internal/handler"
	"github.com/raufhm/vfc/internal/logger"
	"github.com/raufhm/vfc/internal/queue"
	"github.com/raufhm/vfc/internal/repository"
	"github.com/raufhm/vfc/internal/service"
	"github.com/raufhm/vfc/internal/worker"
	"go.uber.org/zap"
)

func main() {
	log, err := logger.New(true)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	log.Info("Starting Product Update Service")

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration", zap.Error(err))
	}

	log.Info("Configuration loaded",
		zap.String("server_port", cfg.Server.Port),
		zap.Int("worker_count", cfg.Worker.Count),
		zap.Int("queue_buffer_size", cfg.Queue.BufferSize))

	repo := repository.NewInMemoryRepository()
	log.Info("Repository initialized")

	q := queue.NewInMemoryQueue(cfg.Queue.BufferSize, log)
	if err := q.Connect(); err != nil {
		log.Fatal("Failed to connect to queue", zap.Error(err))
	}

	svc := service.NewProductService(repo, q)
	log.Info("Service initialized")

	pool := worker.NewPool(cfg.Worker.Count, q, repo, log)
	pool.Start()

	productHandler := handler.NewProductHandler(svc, log)
	router := handler.SetupRouter(productHandler, log)

	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	go func() {
		log.Info("Server starting", zap.String("port", cfg.Server.Port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool.Stop()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown", zap.Error(err))
	}

	if err := q.Close(); err != nil {
		log.Error("Error closing queue", zap.Error(err))
	}

	if err := repo.Close(); err != nil {
		log.Error("Error closing repository", zap.Error(err))
	}

	log.Info("Server stopped gracefully")
}
