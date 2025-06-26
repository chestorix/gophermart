package main

import (
	"context"
	"github.com/chestorix/gophermart/internal/api"
	"github.com/chestorix/gophermart/internal/config"
	"github.com/chestorix/gophermart/internal/repository"
	"github.com/chestorix/gophermart/internal/service"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	jwtSecret := "test"
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	cfg := config.Load()
	storage, err := repository.NewPostgres(cfg.DbURL)
	if err != nil {
		logger.Fatal(err)
	}
	service := service.NewService(storage, logger, jwtSecret, cfg.AccSysAddr)
	server := api.NewServer(cfg, service, logger)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		ticker := time.NewTicker(time.Second * 10)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := service.ProcessOrders(context.Background()); err != nil {
					logger.Errorf("order processing failed: %v", err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Info("Server exiting")
}
