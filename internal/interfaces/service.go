package interfaces

import (
	"context"
	"github.com/chestorix/gophermart/internal/models"
)

type Service interface {
	Test() string

	Register(ctx context.Context, login, password string) (string, error)
	Login(ctx context.Context, login, password string) (string, error)
	GetUserByLogin(ctx context.Context, login string) (models.User, error)

	UploadOrder(ctx context.Context, userID int, orderNumber string) error
	GetUserOrders(ctx context.Context, userID int) ([]models.Order, error)

	Withdraw(ctx context.Context, userID int, orderNumber string, sum float64) error
	GetUserWithdrawals(ctx context.Context, userID int) ([]models.Withdrawal, error)
	GetUserBalance(ctx context.Context, userID int) (current, withdrawn float64, err error)

	ValidateToken(tokenString string) (string, error)
}
