package api

import (
	"encoding/json"
	"fmt"
	"github.com/chestorix/gophermart/internal/interfaces"
	"github.com/chestorix/gophermart/internal/service"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
)

type Handler struct {
	service interfaces.Service
	logger  *logrus.Logger
	dbURL   string
}

func NewHandler(service interfaces.Service, logger *logrus.Logger, dbURL string) *Handler {
	return &Handler{service: service,
		logger: logger,
		dbURL:  dbURL,
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
	if req.Login == "" || req.Password == "" {
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

	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request format", http.StatusBadRequest)
		return
	}
	if req.Login == "" || req.Password == "" {
		http.Error(w, "invalid request format", http.StatusBadRequest)
		return
	}

	token, err := h.service.Login(r.Context(), req.Login, req.Password)
	if err != nil {
		switch err {
		case service.ErrInvalidCredentials:
			http.Error(w, err.Error(), http.StatusUnauthorized)
		default:
			h.logger.Errorf("login failed: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) UploadOrder(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "cannot read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	orderNumber := string(body)
	if orderNumber == "" {
		http.Error(w, "empty order number", http.StatusBadRequest)
		return
	}

	err = h.service.UploadOrder(r.Context(), userID, orderNumber)
	switch err {
	case nil:
		w.WriteHeader(http.StatusAccepted)
	case service.ErrOrderAlreadyUploadedByUser:
		w.WriteHeader(http.StatusOK)
	case service.ErrOrderAlreadyUploadedByAnotherUser:
		http.Error(w, err.Error(), http.StatusConflict)
	case service.ErrInvalidOrderNumber:
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
	default:
		h.logger.Errorf("upload order failed: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
