package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/chestorix/gophermart/internal/models"
	_ "github.com/jackc/pgx/v5/stdlib"
	"time"
)

type Postgres struct {
	db    *sql.DB
	dbURL string
}

func NewPostgres(dsn string) (*Postgres, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}
	return &Postgres{
		db:    db,
		dbURL: dsn,
	}, nil
}

func createTables(db *sql.DB) error {
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    login VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
		
CREATE TABLE IF NOT EXISTS orders (
    number VARCHAR(255) PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    status VARCHAR(50) NOT NULL,
    accrual NUMERIC(10, 2) DEFAULT 0,
    uploaded_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS withdrawals (
    order_number VARCHAR(255) PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    sum NUMERIC(10, 2) NOT NULL,
    processed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
	`)
	return err
}

func (p *Postgres) Test() string {
	return "test"
}

func (p *Postgres) CreateUser(ctx context.Context, user models.User) error {
	query := `INSERT INTO users (login, password_hash) VALUES ($1, $2) RETURNING id`

	var id int
	err := p.db.QueryRowContext(ctx, query, user.Login, user.PasswordHash).Scan(&id)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}
	return nil
}

func (p *Postgres) GetUserByLogin(ctx context.Context, login string) (models.User, error) {
	query := `SELECT id, login, password_hash, created_at FROM users WHERE login = $1`
	var user models.User
	err := p.db.QueryRowContext(ctx, query, login).Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, fmt.Errorf("user not found")
		}
		return models.User{}, err
	}
	return user, nil
}

func (p *Postgres) CreateOrder(ctx context.Context, order models.Order) error {
	query := `
		INSERT INTO orders (number, user_id, status, accrual)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (number) DO NOTHING
	`

	_, err := p.db.ExecContext(ctx, query,
		order.Number,
		order.UserID,
		order.Status,
		order.Accrual,
	)

	return err
}

func (p *Postgres) GetOrderByNumber(ctx context.Context, number string) (models.Order, error) {
	query := `
		SELECT number, user_id, status, accrual, uploaded_at 
		FROM orders 
		WHERE number = $1
	`
	var order models.Order
	err := p.db.QueryRowContext(ctx, query, number).Scan(
		&order.Number,
		&order.UserID,
		&order.Status,
		&order.Accrual,
		&order.UploadedAt)
	if err != nil {
		return models.Order{}, err
	}
	return order, nil
}

func (p *Postgres) GetOrdersByUserID(ctx context.Context, userID int) ([]models.Order, error) {
	query := `
		SELECT number, status, accrual, uploaded_at 
		FROM orders 
		WHERE user_id = $1
		ORDER BY uploaded_at DESC
	`
	rows, err := p.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var orders []models.Order
	for rows.Next() {
		var order models.Order
		if err := rows.Scan(
			&order.Number,
			&order.Status,
			&order.Accrual,
			&order.UploadedAt); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return orders, nil
}

func (p *Postgres) UpdateOrder(ctx context.Context, order models.Order) error {
	query := `
		UPDATE orders 
		SET status = $1, accrual = $2, updated_at = NOW() 
		WHERE number = $3
	`
	_, err := p.db.ExecContext(ctx, query,
		order.Status,
		order.Accrual,
		order.Number)
	return err
}

func (p *Postgres) CreateWithdrawal(ctx context.Context, withdrawal models.Withdrawal) error {
	query := `
		INSERT INTO withdrawals (order_number, user_id, sum)
		VALUES ($1, $2, $3)
	`
	_, err := p.db.ExecContext(ctx, query,
		withdrawal.Order,
		withdrawal.UserID,
		withdrawal.Sum,
	)

	return err

}
func (p *Postgres) GetWithdrawalsByUserID(ctx context.Context, userID int) ([]models.Withdrawal, error) {
	query := `
		SELECT order_number, sum, processed_at 
		FROM withdrawals 
		WHERE user_id = $1
		ORDER BY processed_at DESC
	`
	rows, err := p.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var withdrawals []models.Withdrawal
	for rows.Next() {
		var w models.Withdrawal
		if err := rows.Scan(
			&w.Order,
			&w.Sum,
			&w.ProcessedAt,
		); err != nil {
			return nil, err
		}
		withdrawals = append(withdrawals, w)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return withdrawals, nil
}

func (p *Postgres) GetUserBalance(ctx context.Context, userID int) (current, withdrawn float64, err error) {
	queryCurrent := `
		SELECT COALESCE(SUM(accrual), 0)
		FROM orders 
		WHERE user_id = $1 AND status = $2
	`
	err = p.db.QueryRowContext(ctx, queryCurrent, userID, models.OrderStatusProcessed).Scan(&current)
	if err != nil {
		return 0, 0, err
	}
	queryWithdrawn := `
		SELECT COALESCE(SUM(sum), 0)
		FROM withdrawals 
		WHERE user_id = $1
	`
	err = p.db.QueryRowContext(ctx, queryWithdrawn, userID).Scan(&withdrawn)
	if err != nil {
		return 0, 0, err
	}

	return current - withdrawn, withdrawn, nil

}
