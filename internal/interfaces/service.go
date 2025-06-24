package interfaces

import (
	"context"
	"github.com/chestorix/gophermart/internal/models"
)

type Service interface {
	Test() string

	Register(ctx context.Context, login, password string) (string, error)
	Login(ctx context.Context, login, password string) (string, error)
	ValidateToken(tokenString string) (string, error)
	UploadOrder(ctx context.Context, userID, orderNumber string) error

	CreateUser(ctx context.Context, user models.User) error
	GetUserByLogin(ctx context.Context, login string) (models.User, error)

	CreateOrder(ctx context.Context, order models.Order) error
	GetOrderByNumber(ctx context.Context, number string) (models.Order, error)
	GetOrdersByUserID(ctx context.Context, userID int) ([]models.Order, error)
	UpdateOrder(ctx context.Context, order models.Order) error

	CreateWithdrawal(ctx context.Context, withdrawal models.Withdrawal) error
	GetWithdrawalsByUserID(ctx context.Context, userID int) ([]models.Withdrawal, error)
	GetUserBalance(ctx context.Context, userID string) (current, withdrawn float64, err error)
}
