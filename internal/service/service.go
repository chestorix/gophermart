package service

import (
	"context"
	"errors"
	"github.com/chestorix/gophermart/internal/interfaces"
	"github.com/chestorix/gophermart/internal/models"
	"github.com/golang-jwt/jwt/v4"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"time"
)

var (
	ErrUserAlreadyExists                 = errors.New("user already exists")
	ErrInvalidCredentials                = errors.New("invalid credentials")
	ErrOrderAlreadyUploadedByUser        = errors.New("order already uploaded by user")
	ErrOrderAlreadyUploadedByAnotherUser = errors.New("order already uploaded by another user")
	ErrInvalidOrderNumber                = errors.New("invalid order number")
	ErrInsufficientFunds                 = errors.New("insufficient funds")
)

type Service struct {
	repo      interfaces.Repository
	logger    *logrus.Logger
	jwtSecret string
}

func NewService(repo interfaces.Repository, logger *logrus.Logger, jwtSecret string) *Service {
	return &Service{repo: repo, logger: logger}
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

func (s *Service) CreateUser(ctx context.Context, user models.User) error {
	return s.repo.CreateUser(ctx, user)
}

func (s *Service) GetUserByLogin(ctx context.Context, login string) (models.User, error) {
	return s.repo.GetUserByLogin(ctx, login)
}

func (s *Service) CreateOrder(ctx context.Context, order models.Order) error {
	return s.repo.CreateOrder(ctx, order)
}

func (s *Service) GetOrderByNumber(ctx context.Context, number string) (models.Order, error) {
	return s.repo.GetOrderByNumber(ctx, number)
}
func (s *Service) GetOrdersByUserID(ctx context.Context, userID int) ([]models.Order, error) {
	return s.repo.GetOrdersByUserID(ctx, userID)
}

func (s *Service) UpdateOrder(ctx context.Context, order models.Order) error {
	return s.repo.UpdateOrder(ctx, order)
}
func (s *Service) CreateWithdrawal(ctx context.Context, withdrawal models.Withdrawal) error {
	return s.repo.CreateWithdrawal(ctx, withdrawal)
}
func (s *Service) GetWithdrawalsByUserID(ctx context.Context, userID int) ([]models.Withdrawal, error) {
	return s.repo.GetWithdrawalsByUserID(ctx, userID)
}

func (s *Service) GetUserBalance(ctx context.Context, userID string) (current, withdrawn float64, err error) {
	return s.repo.GetUserBalance(ctx, userID)
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

func (s *Service) UploadOrder(ctx context.Context, userID, orderNumber string) error {
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
		Status:  "NEW",
		Accrual: 0,
	}

	return s.repo.CreateOrder(ctx, order)
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
