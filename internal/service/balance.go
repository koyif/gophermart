package service

import (
	"github.com/koyif/gophermart/internal/domain"
)

type balanceRepository interface {
	Balance(userID int64) (*domain.Balance, error)
}

type withdrawalRepository interface {
	Withdrawals(userID int64) ([]domain.Withdrawal, error)
	Withdraw(orderID string, amount float64, userID int64) error
}

type BalanceService struct {
	balanceRepo    balanceRepository
	withdrawalRepo withdrawalRepository
}

func NewBalanceService(balanceRepo balanceRepository, withdrawalRepo withdrawalRepository) *BalanceService {
	return &BalanceService{
		balanceRepo:    balanceRepo,
		withdrawalRepo: withdrawalRepo,
	}
}

func (b BalanceService) Balance(userID int64) (*domain.Balance, error) {
	return b.balanceRepo.Balance(userID)
}

func (b BalanceService) Withdraw(orderNumber string, sum float64, userID int64) error {
	return b.withdrawalRepo.Withdraw(orderNumber, sum, userID)
}

func (b BalanceService) Withdrawals(userID int64) ([]domain.Withdrawal, error) {
	return b.withdrawalRepo.Withdrawals(userID)
}
