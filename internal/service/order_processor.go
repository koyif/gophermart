package service

import (
	"context"
	"github.com/koyif/gophermart/internal/domain"
	"github.com/koyif/gophermart/pkg/logger"
	"sync"
	"time"
)

type orderProcessorRepository interface {
	FetchPendingOrders() ([]domain.Order, error)
	UpdateOrderStatus(orderID int64, status string, accrual *float64) error
}

type userRepository interface {
	UpdateUserBalance(userID int64, amount *float64) error
}

type OrderProcessor struct {
	orderRepo orderProcessorRepository
	userRepo  userRepository
	mu        *sync.RWMutex
}

func NewOrderProcessor(orderRepo orderProcessorRepository, userRepo userRepository) *OrderProcessor {
	return &OrderProcessor{
		orderRepo: orderRepo,
		userRepo:  userRepo,
		mu:        &sync.RWMutex{},
	}
}

func (p *OrderProcessor) ExtractOrders(ctx context.Context) <-chan domain.Order {
	inputCh := make(chan domain.Order, 1024)
	timer := time.NewTicker(5 * time.Second)
	go func() {
		for {
			select {
			case <-ctx.Done():
				close(inputCh)
				return
			case <-timer.C:
				p.mu.RLock()
				orders, err := p.orderRepo.FetchPendingOrders()
				if err != nil {
					logger.Log.Error("error while fetching pending orders", logger.Error(err))
					p.mu.Unlock()
					continue
				}
				for _, order := range orders {
					inputCh <- order
				}
				p.mu.RUnlock()
			}
		}
	}()

	return inputCh
}

func (p *OrderProcessor) UpdateOrders(ctx context.Context, ordersCh <-chan domain.Order) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case order := <-ordersCh:
				p.mu.Lock()
				err := p.orderRepo.UpdateOrderStatus(order.ID, order.Status, order.Accrual)
				if err != nil {
					p.mu.Unlock()
					logger.Log.Error("error while updating order status", logger.Error(err))
					continue
				}
				err = p.userRepo.UpdateUserBalance(order.UserID, order.Accrual)
				if err != nil {
					p.mu.Unlock()
					logger.Log.Error("error while updating user balance", logger.Error(err))
					return
				}
				p.mu.Unlock()
			}
		}
	}()
}
