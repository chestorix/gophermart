package service

import (
	"context"
	"fmt"
	e "github.com/chestorix/gophermart/internal/errors"
	"github.com/chestorix/gophermart/internal/interfaces"
	"github.com/chestorix/gophermart/internal/models"
	"github.com/go-resty/resty/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"
)

/*var (
	ErrUserAlreadyExists                 = errors.New("user already exists")
	ErrInvalidCredentials                = errors.New("invalid credentials")
	ErrOrderAlreadyUploadedByUser        = errors.New("order already uploaded by user")
	ErrOrderAlreadyUploadedByAnotherUser = errors.New("order already uploaded by another user")
	ErrInvalidOrderNumber                = errors.New("invalid order number")
	ErrInsufficientFunds                 = errors.New("insufficient funds")
	ErrOrderNotRegistered                = errors.New("order not registered")
)*/

type Service struct {
	httpClient *resty.Client
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
		httpClient: resty.New(),
		repo:       repo,
		logger:     logger,
		jwtSecret:  jwtSecret,
		accSysAddr: AccSysAddr,
	}
}

func (s *Service) Register(ctx context.Context, login, password string) (string, error) {

	_, err := s.repo.GetUserByLogin(ctx, login)
	if err == nil {
		return "", e.ErrUserAlreadyExists
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
		return "", e.ErrInvalidCredentials
	}
	if err := s.comparePassword(user.PasswordHash, password); err != nil {
		return "", e.ErrInvalidCredentials
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
		return e.ErrInvalidOrderNumber
	}

	existingOrder, err := s.repo.GetOrderByNumber(ctx, orderNumber)
	if err == nil {
		if existingOrder.UserID == userID {
			return e.ErrOrderAlreadyUploadedByUser
		}
		return e.ErrOrderAlreadyUploadedByAnotherUser
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
		return e.ErrInsufficientFunds
	}

	if !isValidLuhn(orderNumber) {
		return e.ErrInvalidOrderNumber
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
		maxRetries := 3
		for i := 0; i < maxRetries; i++ {
			accrualResp, err := s.GetAccrual(ctx, order.Number)
			if err != nil {
				if i == maxRetries-1 {
					s.logger.Errorf("failed to get accrual for order %s after %d retries: %v",
						order.Number, maxRetries, err)

					order.Status = models.OrderStatusInvalid
					break
				}
				time.Sleep(time.Second * time.Duration(i+1))
				continue
			}

			switch accrualResp.Status {
			case models.AccrualStatusRegistered:
				order.Status = models.OrderStatusNew
			case models.AccrualStatusProcessing:
				order.Status = models.OrderStatusProcessing
			case models.AccrualStatusProcessed:
				order.Status = models.OrderStatusProcessed
				order.Accrual = accrualResp.Accrual
			case models.AccrualStatusInvalid:
				order.Status = models.OrderStatusInvalid
			}
			break
		}

		if err := s.repo.UpdateOrder(ctx, order); err != nil {
			s.logger.Errorf("failed to update order %s: %v", order.Number, err)

			time.AfterFunc(5*time.Minute, func() {
				s.ProcessOrders(context.Background())
			})
		}
	}

	return nil
}

func (s *Service) GetAccrual(ctx context.Context, orderNumber string) (AccrualResponse, error) {

	url := fmt.Sprintf("%s/api/orders/%s", s.accSysAddr, orderNumber)

	var accrualResp AccrualResponse
	resp, err := s.httpClient.R().
		SetContext(ctx).
		SetResult(&accrualResp).
		Get(url)

	if err != nil {
		return AccrualResponse{}, err
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return accrualResp, nil
	case http.StatusNoContent:
		return AccrualResponse{}, e.ErrOrderNotRegistered
	case http.StatusTooManyRequests:
		retryAfter := resp.Header().Get("Retry-After")
		return AccrualResponse{}, fmt.Errorf("rate limit exceeded, retry after %s", retryAfter)
	default:
		return AccrualResponse{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode())
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
			return nil, e.ErrUnexpectedSignMethod
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		login, ok := claims["login"].(string)
		if !ok {
			return "", e.ErrInvalidTokenClaim
		}
		return login, nil
	}

	return "", e.ErrInvalidToken
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
