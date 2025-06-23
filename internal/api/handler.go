package api

import (
	"encoding/json"
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

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request format", http.StatusBadRequest)
		return
	}
	token, err := h.service.Register(r.Context(), req.Login, req.Password)
	if err != nil {
		switch err {
		case service.ErrUserAlreadyExists:
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			h.logger.Errorf("registration failed: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}
}
