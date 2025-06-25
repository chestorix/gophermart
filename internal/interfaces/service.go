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
	UploadOrder(ctx context.Context, userID int, orderNumber string) error

	CreateUser(ctx context.Context, user models.User) error
	GetUserByLogin(ctx context.Context, login string) (models.User, error)

	CreateOrder(ctx context.Context, order models.Order) error
	GetOrderByNumber(ctx context.Context, number string) (models.Order, error)
	GetOrdersByUserID(ctx context.Context, userID int) ([]models.Order, error)
	UpdateOrder(ctx context.Context, order models.Order) error
	GetUserOrders(ctx context.Context, userID int) ([]models.Order, error)

	CreateWithdrawal(ctx context.Context, withdrawal models.Withdrawal) error
	Withdraw(ctx context.Context, userID int, orderNumber string, sum float64) error
	GetUserWithdrawals(ctx context.Context, userID int) ([]models.Withdrawal, error)
	GetWithdrawalsByUserID(ctx context.Context, userID int) ([]models.Withdrawal, error)
	GetUserBalance(ctx context.Context, userID int) (current, withdrawn float64, err error)
}
