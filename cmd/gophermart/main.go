package main

import (
	"github.com/chestorix/gophermart/internal/api"
	"github.com/chestorix/gophermart/internal/config"
	"github.com/chestorix/gophermart/internal/repository"
	"github.com/chestorix/gophermart/internal/service"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	cfg := config.Load()
	storage, err := repository.NewPostgres(cfg.DbURL)
	if err != nil {
		logger.Fatal(err)
	}
	service := service.NewService(storage)
	server := api.NewServer(cfg, service, logger)
	if err := server.Start(); err != nil {
		logger.WithError(err).Fatal("Server failed")
	}

}
