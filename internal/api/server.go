package api

import (
	"context"
	"github.com/chestorix/gophermart/internal/config"
	"github.com/chestorix/gophermart/internal/interfaces"
	"github.com/sirupsen/logrus"
	"net/http"
)

type Server struct {
	cfg     *config.ServerConfig
	router  *Router
	service interfaces.Service
	server  *http.Server
	logger  *logrus.Logger
}

func NewServer(cfg *config.ServerConfig, service interfaces.Service, logger *logrus.Logger) *Server {
	return &Server{
		cfg:     cfg,
		service: service,
		router:  NewRouter(logger),
		logger:  logger,
	}
}

func (s *Server) Start() error {
	s.logger.Info("Starting server...")
	s.router.SetupRoutes(NewHandler(s.service, s.cfg.DbURL))
	httpServer := &http.Server{
		Addr:    s.cfg.RunAddress,
		Handler: s.router,
	}
	s.logger.Infoln("Server listened address: ", s.cfg.RunAddress)
	return httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
