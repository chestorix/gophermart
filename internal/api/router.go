package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
)

type Router struct {
	chi.Router
	logger *logrus.Logger
}

func NewRouter(logger *logrus.Logger) *Router {
	r := chi.NewRouter()

	return &Router{
		Router: r,
		logger: logger,
	}
}
func (r *Router) SetupRoutes(handler *Handler) {
	r.Route("/", func(r chi.Router) {
		r.Get("/", handler.GetTest)
		r.Post("/api/user/register", handler.Register)
	})
}
