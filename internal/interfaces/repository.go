package interfaces

import (
	"context"
	"github.com/chestorix/gophermart/internal/models"
)

type Repository interface {
	Test() string
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
