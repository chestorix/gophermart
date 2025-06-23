package api

import (
	"fmt"
	"github.com/chestorix/gophermart/internal/interfaces"
	"net/http"
)

type Handler struct {
	service interfaces.Service
	dbURL   string
}

func NewHandler(service interfaces.Service, dbURL string) *Handler {
	return &Handler{service: service,
		dbURL: dbURL,
	}
}

func (h *Handler) GetTest(w http.ResponseWriter, r *http.Request) {
	test := h.service.Test()
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, test)
}
