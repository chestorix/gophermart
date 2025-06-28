package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/chestorix/gophermart/internal/interfaces"
	"github.com/chestorix/gophermart/internal/models"
	"github.com/golang-jwt/jwt/v4"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"
)

var (
	ErrUserAlreadyExists                 = errors.New("user already exists")
	ErrInvalidCredentials                = errors.New("invalid credentials")
	ErrOrderAlreadyUploadedByUser        = errors.New("order already uploaded by user")
	ErrOrderAlreadyUploadedByAnotherUser = errors.New("order already uploaded by another user")
	ErrInvalidOrderNumber                = errors.New("invalid order number")
	ErrInsufficientFunds                 = errors.New("insufficient funds")
	ErrOrderNotRegistered                = errors.New("order not registered")
)

type Service struct {
	httpClient *http.Client
	repo       interfaces.Repository
	logger     *logrus.Logger
	jwtSecret  string
	accSysAddr string
}

type AccrualResponse struct {
	Order   string               `json:"order"`
	Status  models.AccrualStatus `json:"status"`
	Accrual float64              `json:"accrual,omitempty"`
}

func NewService(repo interfaces.Repository, logger *logrus.Logger, jwtSecret string, AccSysAddr string) *Service {
	return &Service{
		httpClient: &http.Client{},
		repo:       repo,
		logger:     logger,
		jwtSecret:  jwtSecret,
		accSysAddr: AccSysAddr,
	}
}

func (s *Service) Register(ctx context.Context, login, password string) (string, error) {

	_, err := s.repo.GetUserByLogin(ctx, login)
	if err == nil {
		return "", ErrUserAlreadyExists
	}

	hashedPassword, err := s.hashPassword(password)
	if err != nil {
		return "", err
	}

	user := models.User{
		Login:        login,
		PasswordHash: hashedPassword,
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return "", err
	}

	token, err := s.generateToken(login)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *Service) Login(ctx context.Context, login, password string) (string, error) {
	user, err := s.repo.GetUserByLogin(ctx, login)
	if err != nil {
		return "", ErrInvalidCredentials
	}
	if err := s.comparePassword(user.PasswordHash, password); err != nil {
		return "", ErrInvalidCredentials
	}

	return s.generateToken(login)
}

func (s *Service) Test() string {
	return s.repo.Test()
}

func (s *Service) GetUserByLogin(ctx context.Context, login string) (models.User, error) {
	return s.repo.GetUserByLogin(ctx, login)
}

func (s *Service) GetUserBalance(ctx context.Context, userID int) (current, withdrawn float64, err error) {
	return s.repo.GetUserBalance(ctx, userID)
}

func (s *Service) UploadOrder(ctx context.Context, userID int, orderNumber string) error {
	if !isValidLuhn(orderNumber) {
		return ErrInvalidOrderNumber
	}

	existingOrder, err := s.repo.GetOrderByNumber(ctx, orderNumber)
	if err == nil {
		if existingOrder.UserID == userID {
			return ErrOrderAlreadyUploadedByUser
		}
		return ErrOrderAlreadyUploadedByAnotherUser
	}

	order := models.Order{
		Number:  orderNumber,
		UserID:  userID,
		Status:  models.OrderStatusNew,
		Accrual: 0,
	}

	return s.repo.CreateOrder(ctx, order)
}

func (s *Service) GetUserOrders(ctx context.Context, userID int) ([]models.Order, error) {
	return s.repo.GetOrdersByUserID(ctx, userID)
}

func (s *Service) Withdraw(ctx context.Context, userID int, orderNumber string, sum float64) error {
	current, _, err := s.repo.GetUserBalance(ctx, userID)
	if err != nil {
		return err
	}

	if current < sum {
		return ErrInsufficientFunds
	}

	if !isValidLuhn(orderNumber) {
		return ErrInvalidOrderNumber
	}

	withdrawal := models.Withdrawal{
		Order:  orderNumber,
		UserID: userID,
		Sum:    sum,
	}

	return s.repo.CreateWithdrawal(ctx, withdrawal)
}

func (s *Service) GetUserWithdrawals(ctx context.Context, userID int) ([]models.Withdrawal, error) {
	return s.repo.GetWithdrawalsByUserID(ctx, userID)
}

func (s *Service) ProcessOrders(ctx context.Context) error {
	orders, err := s.repo.GetOrdersToProcess(ctx, 10)
	if err != nil {
		return err
	}

	for _, order := range orders {
		accrualResp, err := s.GetAccrual(ctx, order.Number)
		if err != nil {
			s.logger.Errorf("failed to get accrual for order %s: %v", order.Number, err)
			continue
		}

		switch accrualResp.Status {
		case models.AccrualStatusRegistered:
			order.Status = models.OrderStatusNew
			order.Accrual = accrualResp.Accrual
		case models.AccrualStatusProcessing:
			order.Status = models.OrderStatusProcessing
			order.Accrual = accrualResp.Accrual
		case models.AccrualStatusProcessed:
			order.Status = models.OrderStatusProcessed
			order.Accrual = accrualResp.Accrual
		case models.AccrualStatusInvalid:
			order.Status = models.OrderStatusInvalid
			order.Accrual = accrualResp.Accrual
		}

		if err := s.repo.UpdateOrder(ctx, order); err != nil {
			s.logger.Errorf("failed to update order %s: %v", order.Number, err)
		}
	}

	return nil
}

func (s *Service) GetAccrual(ctx context.Context, orderNumber string) (AccrualResponse, error) {
	url := fmt.Sprintf("%s/api/orders/%s", s.accSysAddr, orderNumber)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return AccrualResponse{}, err
	}

	resp, err := s.httpClient.Do(req)
	s.logger.Infoln(url)
	if err != nil {
		return AccrualResponse{}, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var accrualResp AccrualResponse
		if err := json.NewDecoder(resp.Body).Decode(&accrualResp); err != nil {
			return AccrualResponse{}, err
		}
		return accrualResp, nil
	case http.StatusNoContent:
		return AccrualResponse{}, ErrOrderNotRegistered
	case http.StatusTooManyRequests:
		retryAfter := resp.Header.Get("Retry-After")
		return AccrualResponse{}, fmt.Errorf("rate limit exceeded, retry after %s", retryAfter)
	default:
		return AccrualResponse{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

func (s *Service) hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func (s *Service) generateToken(login string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"login": login,
		"exp":   time.Now().Add(time.Hour * 24).Unix(),
	})
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *Service) comparePassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func (s *Service) ValidateToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		login, ok := claims["login"].(string)
		if !ok {
			return "", errors.New("invalid token claims")
		}
		return login, nil
	}

	return "", errors.New("invalid token")
}

func isValidLuhn(number string) bool {
	// Реализация алгоритма Луна
	sum := 0
	alternate := false

	for i := len(number) - 1; i >= 0; i-- {
		digit := int(number[i] - '0')
		if digit < 0 || digit > 9 {
			return false
		}

		if alternate {
			digit *= 2
			if digit > 9 {
				digit = (digit % 10) + 1
			}
		}

		sum += digit
		alternate = !alternate
	}

	return sum%10 == 0
}
