package service

import "github.com/koyif/gophermart/internal/domain"

type OrderRepository interface {
	CreateOrder(orderID string, userID int64) error
	Orders(userID int64) ([]domain.Order, error)
}

type OrderService struct {
	repo OrderRepository
}

func NewOrderService(repo OrderRepository) *OrderService {
	return &OrderService{
		repo: repo,
	}
}

func (s *OrderService) Create(orderID string, userID int64) error {
	return s.repo.CreateOrder(orderID, userID)
}

func (s *OrderService) Orders(userID int64) ([]domain.Order, error) {
	orders, err := s.repo.Orders(userID)
	if err != nil {
		return nil, err
	}

	return orders, nil
}
