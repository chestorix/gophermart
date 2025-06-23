package service

import (
	"context"
	"errors"
	"github.com/chestorix/gophermart/internal/interfaces"
	"github.com/chestorix/gophermart/internal/models"
)

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type Service struct {
	repo interfaces.Repository
}

func NewService(repo interfaces.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Register(ctx context.Context, login, password string) (string, error) {

	_, err := s.repo.GetUserByLogin(ctx, login)
	if err == nil {
		return "", ErrUserAlreadyExists
	}

	hashedPassword, err := uc.authService.HashPassword(password)
	if err != nil {
		return "", err
	}

	// Создание пользователя
	user := models.User{
		Login:        login,
		PasswordHash: hashedPassword,
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return "", err
	}

	// Генерация токена
	token, err := uc.authService.GenerateToken(login)
	if err != nil {
		return "", err
	}

	return token, nil
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
