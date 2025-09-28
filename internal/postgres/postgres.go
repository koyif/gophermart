package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/koyif/gophermart/internal/domain"
	"github.com/koyif/gophermart/pkg/logger"
)

const transactionRollbackError = "error rolling back transaction"

type Postgres struct {
	DB *sql.DB
}

func New(db *sql.DB) *Postgres {
	return &Postgres{DB: db}
}

func (p *Postgres) Close() error {
	return p.DB.Close()
}

func (p *Postgres) CreateUser(login, hashedPassword string) (int64, error) {
	var id int64
	err := p.DB.QueryRow("INSERT INTO users (login, password) VALUES ($1, $2) RETURNING id", login, hashedPassword).
		Scan(&id)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			logger.Log.Warn("user already exists", logger.String("login", login))
			return 0, domain.ErrUserExists
		}
		return 0, fmt.Errorf("error creating user: %w", err)
	}

	return id, nil
}

func (p *Postgres) User(login string) (*domain.User, error) {
	row := p.DB.QueryRow("SELECT id, login, password, registered_at FROM users WHERE login = $1", login)

	var user domain.User
	err := row.Scan(&user.ID, &user.Login, &user.Password, &user.RegisteredAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrIncorrectCredentials
		}
		return nil, fmt.Errorf("error fetching user: %w", err)
	}

	return &user, nil
}

func (p *Postgres) CreateOrder(orderNumber string, userID int64) error {
	tx, err := p.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}

	var order domain.Order
	err = tx.QueryRow("SELECT number, user_id FROM orders WHERE number = $1", orderNumber).
		Scan(&order.Number, &order.UserID)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		rollback(tx)
		return fmt.Errorf("error fetching order: %w", err)
	}

	if order.UserID != 0 && order.UserID != userID {
		logger.Log.Warn(
			"order already exists for different user",
			logger.String("number", orderNumber),
			logger.Int64("existing_user_id", order.UserID),
			logger.Int64("new_user_id", userID),
		)
		rollback(tx)
		return domain.ErrOrderAddedByAnotherUser
	} else if order.UserID != 0 && order.UserID == userID {
		logger.Log.Warn("order already exists", logger.String("number", orderNumber))
		rollback(tx)
		return domain.ErrOrderExists
	}

	_, err = p.DB.Exec("INSERT INTO orders (number, user_id) VALUES ($1, $2)", orderNumber, userID)
	if err != nil {
		rollback(tx)
		return fmt.Errorf("error creating order: %w", err)
	}
	err = tx.Commit()
	if err != nil {
		rollback(tx)
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

func (p *Postgres) Orders(userID int64) ([]domain.Order, error) {
	rows, err := p.DB.Query("SELECT number, user_id, status, accrual, uploaded_at FROM orders WHERE user_id = $1", userID)
	if err != nil {
		return nil, fmt.Errorf("error fetching orders: %w", err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			logger.Log.Error("error closing rows", logger.Error(err))
		}
	}(rows)

	var orders []domain.Order
	for rows.Next() {
		var order domain.Order
		err := rows.Scan(&order.Number, &order.UserID, &order.Status, &order.Accrual, &order.UploadedAt)
		if err != nil {
			return nil, fmt.Errorf("error scanning order: %w", err)
		}
		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over orders: %w", err)
	}

	return orders, nil
}

func (p *Postgres) FetchPendingOrders() ([]domain.Order, error) {
	rows, err := p.DB.Query("SELECT id, number, user_id, status FROM orders WHERE status = 'NEW' OR status = 'PROCESSING'")
	if err != nil {
		return nil, fmt.Errorf("error fetching orders: %w", err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			logger.Log.Error("error closing rows", logger.Error(err))
		}
	}(rows)

	var orders []domain.Order
	for rows.Next() {
		var order domain.Order
		err := rows.Scan(&order.ID, &order.Number, &order.UserID, &order.Status)
		if err != nil {
			return nil, fmt.Errorf("error scanning order: %w", err)
		}
		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over orders: %w", err)
	}

	return orders, nil
}

func (p *Postgres) UpdateOrderStatus(orderID int64, status string, accrual *float64) error {
	_, err := p.DB.Exec("UPDATE orders SET status = $1, accrual = $2 WHERE id = $3", status, accrual, orderID)
	if err != nil {
		return fmt.Errorf("error updating order status: %w", err)
	}

	return nil
}

func (p *Postgres) UpdateUserBalance(userID int64, amount *float64) error {
	if amount == nil {
		return nil
	}
	_, err := p.DB.Exec("UPDATE users SET balance = balance + $1 WHERE id = $2", *amount, userID)
	if err != nil {
		return fmt.Errorf("error updating user balance: %w", err)
	}

	return nil
}

func (p *Postgres) Balance(userID int64) (*domain.Balance, error) {
	var balance domain.Balance
	err := p.DB.QueryRow("SELECT balance, withdrawn FROM users WHERE id = $1", userID).
		Scan(&balance.Current, &balance.Withdrawn)

	if err != nil {
		return nil, fmt.Errorf("error fetching balance: %w", err)
	}

	return &balance, nil
}

func (p *Postgres) Withdrawals(userID int64) ([]domain.Withdrawal, error) {
	rows, err := p.DB.Query("SELECT order_number, amount, processed_at FROM withdrawals WHERE user_id = $1", userID)
	if err != nil {
		return nil, fmt.Errorf("error fetching withdrawals: %w", err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			logger.Log.Error("error closing rows", logger.Error(err))
		}
	}(rows)

	var withdrawals []domain.Withdrawal
	for rows.Next() {
		var withdrawal domain.Withdrawal
		err := rows.Scan(&withdrawal.OrderNumber, &withdrawal.Amount, &withdrawal.ProcessedAt)
		if err != nil {
			return nil, fmt.Errorf("error scanning withdrawal: %w", err)
		}
		withdrawals = append(withdrawals, withdrawal)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over withdrawals: %w", err)
	}

	return withdrawals, nil
}

func (p *Postgres) Withdraw(orderNumber string, amount float64, userID int64) error {
	tx, err := p.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}

	var withdrawal domain.Withdrawal
	err = tx.QueryRow("SELECT user_id, order_number FROM withdrawals WHERE order_number = $1", orderNumber).
		Scan(&withdrawal.UserID, &withdrawal.OrderNumber)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		rollback(tx)
		return fmt.Errorf("error fetching withdrawal: %w", err)
	}

	if withdrawal.UserID != 0 && withdrawal.UserID != userID {
		logger.Log.Warn(
			"withdrawal already exists for different user",
			logger.String("number", orderNumber),
			logger.Int64("existing_user_id", withdrawal.UserID),
			logger.Int64("new_user_id", userID),
		)
		rollback(tx)
		return domain.ErrWithdrawalAddedByAnotherUser
	} else if withdrawal.UserID != 0 && withdrawal.UserID == userID {
		logger.Log.Warn("withdrawal already exists", logger.String("number", orderNumber))
		rollback(tx)
		return domain.ErrWithdrawalExists
	}

	_, err = tx.Exec("INSERT INTO withdrawals (order_number, amount, user_id) VALUES ($1, $2, $3)", orderNumber, amount, userID)
	if err != nil {
		rollback(tx)
		logger.Log.Error("error inserting withdrawal", logger.String("order_id", orderNumber), logger.Float64("amount", amount), logger.Int64("user_id", userID), logger.Error(err))
		return fmt.Errorf("error inserting withdrawal: %w", err)
	}

	result, err := tx.Exec("UPDATE users SET balance = balance - $1, withdrawn = withdrawn + $1 WHERE id = $2 AND balance >= $1", amount, userID)
	if err != nil {
		rollback(tx)
		logger.Log.Error("error updating user balance for withdrawal", logger.Float64("amount", amount), logger.Int64("user_id", userID), logger.Error(err))
		return fmt.Errorf("error updating user balance for withdrawal: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		rollback(tx)
		logger.Log.Error("error checking rows affected for balance update", logger.Float64("amount", amount), logger.Int64("user_id", userID), logger.Error(err))
		return fmt.Errorf("error checking rows affected for balance update: %w", err)
	}
	if rowsAffected == 0 {
		rollback(tx)
		logger.Log.Warn("insufficient funds for withdrawal", logger.Float64("amount", amount), logger.Int64("user_id", userID))
		return domain.ErrInsufficientFunds
	}

	err = tx.Commit()
	if err != nil {
		rollback(tx)
		logger.Log.Error("error committing transaction for withdrawal", logger.Float64("amount", amount), logger.Int64("user_id", userID), logger.Error(err))
		return fmt.Errorf("error committing transaction for withdrawal: %w", err)
	}

	return nil
}

func rollback(tx *sql.Tx) {
	err := tx.Rollback()
	if err != nil {
		logger.Log.Error(transactionRollbackError, logger.Error(err))
	}
}
