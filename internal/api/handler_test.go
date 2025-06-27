package api

import (
	"context"
	"github.com/chestorix/gophermart/internal/models"
	"github.com/chestorix/gophermart/internal/service"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type mockService struct {
	registerFn           func(ctx context.Context, login, password string) (string, error)
	loginFn              func(ctx context.Context, login, password string) (string, error)
	uploadOrderFn        func(ctx context.Context, userID int, orderNumber string) error
	getUserOrdersFn      func(ctx context.Context, userID int) ([]models.Order, error)
	getUserBalanceFn     func(ctx context.Context, userID int) (current, withdrawn float64, err error)
	withdrawFn           func(ctx context.Context, userID int, orderNumber string, sum float64) error
	getUserWithdrawalsFn func(ctx context.Context, userID int) ([]models.Withdrawal, error)
	validateTokenFn      func(tokenString string) (string, error)
	getUserByLoginFn     func(ctx context.Context, login string) (models.User, error)
}

func (m *mockService) Test() string {
	return "test"
}

func (m *mockService) Register(ctx context.Context, login, password string) (string, error) {
	return m.registerFn(ctx, login, password)
}

func (m *mockService) Login(ctx context.Context, login, password string) (string, error) {
	return m.loginFn(ctx, login, password)
}

func (m *mockService) UploadOrder(ctx context.Context, userID int, orderNumber string) error {
	return m.uploadOrderFn(ctx, userID, orderNumber)
}

func (m *mockService) GetUserOrders(ctx context.Context, userID int) ([]models.Order, error) {
	return m.getUserOrdersFn(ctx, userID)
}

func (m *mockService) GetUserBalance(ctx context.Context, userID int) (current, withdrawn float64, err error) {
	return m.getUserBalanceFn(ctx, userID)
}

func (m *mockService) Withdraw(ctx context.Context, userID int, orderNumber string, sum float64) error {
	return m.withdrawFn(ctx, userID, orderNumber, sum)
}

func (m *mockService) GetUserWithdrawals(ctx context.Context, userID int) ([]models.Withdrawal, error) {
	return m.getUserWithdrawalsFn(ctx, userID)
}

func (m *mockService) ValidateToken(tokenString string) (string, error) {
	return m.validateTokenFn(tokenString)
}

func (m *mockService) GetUserByLogin(ctx context.Context, login string) (models.User, error) {
	return m.getUserByLoginFn(ctx, login)
}

func TestHandler_Register(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		mockRegister   func(ctx context.Context, login, password string) (string, error)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "successful registration",
			requestBody: `{"login": "user1", "password": "pass123"}`,
			mockRegister: func(ctx context.Context, login, password string) (string, error) {
				return "token123", nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "user already exists",
			requestBody: `{"login": "user1", "password": "pass123"}`,
			mockRegister: func(ctx context.Context, login, password string) (string, error) {
				return "", service.ErrUserAlreadyExists
			},
			expectedStatus: http.StatusConflict,
			expectedBody:   service.ErrUserAlreadyExists.Error() + "\n",
		},
		{
			name:        "invalid request body",
			requestBody: `invalid json`,
			mockRegister: func(ctx context.Context, login, password string) (string, error) {
				return "", nil
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid request format\n",
		},
		{
			name:        "empty login or password",
			requestBody: `{"login": "", "password": "pass123"}`,
			mockRegister: func(ctx context.Context, login, password string) (string, error) {
				return "", nil
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid request format\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &mockService{
				registerFn: tt.mockRegister,
			}

			handler := NewHandler(service, logrus.New(), "")

			req := httptest.NewRequest("POST", "/api/user/register", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.Register(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.expectedBody != "" {
				body, _ := io.ReadAll(resp.Body)
				if string(body) != tt.expectedBody {
					t.Errorf("expected body %q, got %q", tt.expectedBody, string(body))
				}
			}

			if tt.expectedStatus == http.StatusOK {
				authHeader := resp.Header.Get("Authorization")
				if authHeader != "Bearer token123" {
					t.Errorf("expected Authorization header 'Bearer token123', got %q", authHeader)
				}
			}
		})
	}
}

func TestHandler_Login(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		mockLogin      func(ctx context.Context, login, password string) (string, error)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "successful login",
			requestBody: `{"login": "user1", "password": "pass123"}`,
			mockLogin: func(ctx context.Context, login, password string) (string, error) {
				return "token123", nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "invalid credentials",
			requestBody: `{"login": "user1", "password": "wrongpass"}`,
			mockLogin: func(ctx context.Context, login, password string) (string, error) {
				return "", service.ErrInvalidCredentials
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   service.ErrInvalidCredentials.Error() + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &mockService{
				loginFn: tt.mockLogin,
			}

			handler := NewHandler(service, logrus.New(), "")

			req := httptest.NewRequest("POST", "/api/user/login", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.Login(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.expectedBody != "" {
				body, _ := io.ReadAll(resp.Body)
				if string(body) != tt.expectedBody {
					t.Errorf("expected body %q, got %q", tt.expectedBody, string(body))
				}
			}

			if tt.expectedStatus == http.StatusOK {
				authHeader := resp.Header.Get("Authorization")
				if authHeader != "Bearer token123" {
					t.Errorf("expected Authorization header 'Bearer token123', got %q", authHeader)
				}
			}
		})
	}
}

func TestHandler_UploadOrder(t *testing.T) {
	tests := []struct {
		name            string
		contentType     string
		orderNumber     string
		mockUploadOrder func(ctx context.Context, userID int, orderNumber string) error
		expectedStatus  int
		expectedBody    string
	}{
		{
			name:        "successful upload",
			contentType: "text/plain",
			orderNumber: "1234567890",
			mockUploadOrder: func(ctx context.Context, userID int, orderNumber string) error {
				return nil
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name:        "invalid content type",
			contentType: "application/json",
			orderNumber: "1234567890",
			mockUploadOrder: func(ctx context.Context, userID int, orderNumber string) error {
				return nil
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid request format\n",
		},
		{
			name:        "order already uploaded by user",
			contentType: "text/plain",
			orderNumber: "1234567890",
			mockUploadOrder: func(ctx context.Context, userID int, orderNumber string) error {
				return service.ErrOrderAlreadyUploadedByUser
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "order already uploaded by another user",
			contentType: "text/plain",
			orderNumber: "1234567890",
			mockUploadOrder: func(ctx context.Context, userID int, orderNumber string) error {
				return service.ErrOrderAlreadyUploadedByAnotherUser
			},
			expectedStatus: http.StatusConflict,
			expectedBody:   service.ErrOrderAlreadyUploadedByAnotherUser.Error() + "\n",
		},
		{
			name:        "invalid order number",
			contentType: "text/plain",
			orderNumber: "invalid",
			mockUploadOrder: func(ctx context.Context, userID int, orderNumber string) error {
				return service.ErrInvalidOrderNumber
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   service.ErrInvalidOrderNumber.Error() + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &mockService{
				uploadOrderFn: tt.mockUploadOrder,
				validateTokenFn: func(tokenString string) (string, error) {
					return "user1", nil
				},
				getUserByLoginFn: func(ctx context.Context, login string) (models.User, error) {
					return models.User{ID: 1}, nil
				},
			}

			handler := NewHandler(service, logrus.New(), "")

			req := httptest.NewRequest("POST", "/api/user/orders", strings.NewReader(tt.orderNumber))
			req.Header.Set("Content-Type", tt.contentType)
			// Добавляем userID в контекст как это делает middleware
			ctx := context.WithValue(req.Context(), "userID", 1)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()

			handler.UploadOrder(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.expectedBody != "" {
				body, _ := io.ReadAll(resp.Body)
				if string(body) != tt.expectedBody {
					t.Errorf("expected body %q, got %q", tt.expectedBody, string(body))
				}
			}
		})
	}
}

func TestHandler_GetUserOrders(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name              string
		mockGetUserOrders func(ctx context.Context, userID int) ([]models.Order, error)
		expectedStatus    int
		expectedBody      string
	}{
		{
			name: "successful get orders",
			mockGetUserOrders: func(ctx context.Context, userID int) ([]models.Order, error) {
				return []models.Order{
					{
						Number:     "1234567890",
						Status:     models.OrderStatusProcessed,
						Accrual:    100.5,
						UploadedAt: now,
					},
				}, nil
			},
			expectedStatus: http.StatusOK,
			expectedBody: `[{"number":"1234567890","status":"PROCESSED","accrual":100.5,"uploaded_at":"` + now.Format(time.RFC3339Nano) + `"}]
`,
		},
		{
			name: "no orders",
			mockGetUserOrders: func(ctx context.Context, userID int) ([]models.Order, error) {
				return []models.Order{}, nil
			},
			expectedStatus: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &mockService{
				getUserOrdersFn: tt.mockGetUserOrders,
				validateTokenFn: func(tokenString string) (string, error) {
					return "user1", nil
				},
				getUserByLoginFn: func(ctx context.Context, login string) (models.User, error) {
					return models.User{ID: 1}, nil
				},
			}

			handler := NewHandler(service, logrus.New(), "")

			req := httptest.NewRequest("GET", "/api/user/orders", nil)
			// Добавляем userID в контекст как это делает middleware
			ctx := context.WithValue(req.Context(), "userID", 1)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()

			handler.GetUserOrders(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.expectedBody != "" {
				body, _ := io.ReadAll(resp.Body)
				if string(body) != tt.expectedBody {
					t.Errorf("expected body %q, got %q", tt.expectedBody, string(body))
				}
			}
		})
	}
}
