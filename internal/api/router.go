package api

import (
	//"github.com/chestorix/gophermart/internal/interfaces"
	mw "github.com/chestorix/gophermart/internal/api/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"
)

type Router struct {
	chi.Router
	logger *logrus.Logger
}

func NewRouter(logger *logrus.Logger) *Router {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))

	return &Router{
		Router: r,
		logger: logger,
	}
}
func (r *Router) SetupRoutes(handler *Handler) {
	r.Group(func(r chi.Router) {
		r.Post("/api/user/register", handler.Register)
		r.Post("/api/user/login", handler.Login)
	})

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(mw.Auth(handler.service))

		r.Post("/api/user/orders", handler.UploadOrder)
		r.Get("/api/user/orders", handler.GetUserOrders)
		//	r.Get("/api/user/balance", handler.GetBalance)
		//	r.Post("/api/user/balance/withdraw", handler.Withdraw)
		//	r.Get("/api/user/withdrawals", handler.GetWithdrawals)
	})
}
