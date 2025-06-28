package api

import (
	"encoding/json"
	"fmt"
	"github.com/chestorix/gophermart/internal/api/middleware"
	e "github.com/chestorix/gophermart/internal/errors"
	"github.com/chestorix/gophermart/internal/interfaces"
	"github.com/chestorix/gophermart/internal/models"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"time"
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
		case e.ErrUserAlreadyExists:
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
		case e.ErrInvalidCredentials:
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
	if r.Header.Get("Content-Type") != "text/plain" {
		http.Error(w, "invalid request format", http.StatusBadRequest)
	}
	userID := r.Context().Value(middleware.UserIDKey).(int)

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
	case e.ErrOrderAlreadyUploadedByUser:
		w.WriteHeader(http.StatusOK)
	case e.ErrOrderAlreadyUploadedByAnotherUser:
		http.Error(w, err.Error(), http.StatusConflict)
	case e.ErrInvalidOrderNumber:
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
	default:
		h.logger.Errorf("upload order failed: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
func (h *Handler) GetUserOrders(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)
	orders, err := h.service.GetUserOrders(r.Context(), userID)
	if err != nil {
		h.logger.Errorf("get user orders failed: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	type orderResponse struct {
		Number     string             `json:"number"`
		Status     models.OrderStatus `json:"status"`
		Accrual    float64            `json:"accrual,omitempty"`
		UploadedAt time.Time          `json:"uploaded_at"`
	}

	response := make([]orderResponse, 0, len(orders))
	for _, order := range orders {
		response = append(response, orderResponse{
			Number:     order.Number,
			Status:     order.Status,
			Accrual:    order.Accrual,
			UploadedAt: order.UploadedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Errorf("encode response failed: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

func (h *Handler) GetUserBalance(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)
	current, withdrawn, err := h.service.GetUserBalance(r.Context(), userID)
	if err != nil {
		h.logger.Errorf("get user balance failed: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	response := struct {
		Current   float64 `json:"current"`
		Withdrawn float64 `json:"withdrawn"`
	}{
		Current:   current,
		Withdrawn: withdrawn,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Errorf("encode response failed: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

func (h *Handler) Withdraw(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)

	var req struct {
		Order string  `json:"order"`
		Sum   float64 `json:"sum"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request format", http.StatusBadRequest)
		return
	}

	err := h.service.Withdraw(r.Context(), userID, req.Order, req.Sum)
	switch err {
	case nil:
		w.WriteHeader(http.StatusOK)
	case e.ErrInsufficientFunds:
		http.Error(w, err.Error(), http.StatusPaymentRequired)
	case e.ErrInvalidOrderNumber:
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
	default:
		h.logger.Errorf("withdraw failed: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
func (h *Handler) GetUserWithdrawals(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)

	withdrawals, err := h.service.GetUserWithdrawals(r.Context(), userID)
	if err != nil {
		h.logger.Errorf("get user withdrawals failed: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	type withdrawalResponse struct {
		Order       string    `json:"order"`
		Sum         float64   `json:"sum"`
		ProcessedAt time.Time `json:"processed_at"`
	}

	response := make([]withdrawalResponse, 0, len(withdrawals))
	for _, w := range withdrawals {
		response = append(response, withdrawalResponse{
			Order:       w.Order,
			Sum:         w.Sum,
			ProcessedAt: w.ProcessedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Errorf("encode response failed: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
