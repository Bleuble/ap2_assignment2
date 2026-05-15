package usecase

import (
	"fmt"
	"order-service/internal/domain"
	"sync" // Added for thread-safety
	"time"

	"github.com/google/uuid"
)

type OrderUseCase struct {
	repo          domain.OrderRepository
	paymentClient domain.PaymentClient
	cache         domain.OrderCache
	mu            sync.RWMutex
	subscribers   map[string][]chan string
}

func NewOrderUseCase(repo domain.OrderRepository, paymentClient domain.PaymentClient, cache domain.OrderCache) *OrderUseCase {
	return &OrderUseCase{
		repo:          repo,
		paymentClient: paymentClient,
		cache:         cache,
		subscribers:   make(map[string][]chan string),
	}
}

func (uc *OrderUseCase) Subscribe(orderID string) (chan string, func()) {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	ch := make(chan string, 1)
	uc.subscribers[orderID] = append(uc.subscribers[orderID], ch)

	cleanup := func() {
		uc.mu.Lock()
		defer uc.mu.Unlock()
		subs := uc.subscribers[orderID]
		for i, sub := range subs {
			if sub == ch {
				uc.subscribers[orderID] = append(subs[:i], subs[i+1:]...)
				close(ch)
				break
			}
		}
	}

	return ch, cleanup
}

func (uc *OrderUseCase) notify(orderID string, status string) {
	uc.mu.RLock()
	defer uc.mu.RUnlock()

	if subs, ok := uc.subscribers[orderID]; ok {
		for _, ch := range subs {
			select {
			case ch <- status:
			default:

			}
		}
	}
}

func (uc *OrderUseCase) CreateOrder(customerID, itemName string, amount int64, idempotencyKey string) (*domain.Order, error) {
	if idempotencyKey != "" {
		existingOrder, err := uc.repo.GetByIdempotencyKey(idempotencyKey)
		if err == nil && existingOrder != nil {
			return existingOrder, nil
		}
	}

	order := &domain.Order{
		ID:             uuid.New().String(),
		CustomerID:     customerID,
		ItemName:       itemName,
		Amount:         amount,
		Status:         "Pending",
		IdempotencyKey: idempotencyKey,
		CreatedAt:      time.Now(),
	}

	if err := order.Validate(); err != nil {
		return nil, err
	}

	if err := uc.repo.Create(order); err != nil {
		return nil, fmt.Errorf("failed to save order: %v", err)
	}
	if uc.cache != nil {
		uc.cache.Set(order)
	}
	uc.notify(order.ID, order.Status)

	_, err := uc.paymentClient.AuthorizePayment(order.ID, order.Amount)
	if err != nil {
		order.DanaFailed()
		uc.repo.UpdateStatus(order.ID, order.Status)
		if uc.cache != nil {
			uc.cache.Invalidate(order.ID)
		}
		uc.notify(order.ID, order.Status)
		return order, fmt.Errorf("payment failed: %v", err)
	}

	order.DanaPaid()
	uc.repo.UpdateStatus(order.ID, order.Status)
	if uc.cache != nil {
		uc.cache.Invalidate(order.ID)
	}
	uc.notify(order.ID, order.Status)

	return order, nil
}

func (uc *OrderUseCase) GetOrder(id string) (*domain.Order, error) {
	if uc.cache != nil {
		cachedOrder, err := uc.cache.Get(id)
		if err == nil && cachedOrder != nil {
			return cachedOrder, nil
		}
	}

	order, err := uc.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if uc.cache != nil && order != nil {
		uc.cache.Set(order)
	}

	return order, nil
}

func (uc *OrderUseCase) CancelOrder(id string) error {
	order, err := uc.repo.GetByID(id)
	if err != nil {
		return err
	}

	if err := order.Cancel(); err != nil {
		return err
	}

	err = uc.repo.UpdateStatus(order.ID, order.Status)
	if err == nil {
		if uc.cache != nil {
			uc.cache.Invalidate(order.ID)
		}
		uc.notify(order.ID, order.Status)
	}
	return err
}

func (uc *OrderUseCase) GetOrdersByAmountRange(min, max int64) ([]*domain.Order, error) {
	if min < 0 {
		return nil, fmt.Errorf("min amount must be at least 0")
	}
	if max > 100000 {
		return nil, fmt.Errorf("max amount must not exceed 100000")
	}
	if min >= max {
		return nil, fmt.Errorf("min amount must be less than max amount")
	}

	return uc.repo.GetByAmountRange(min, max)
}

func (uc *OrderUseCase) ListPayments(status string) (interface{}, error) {
	return uc.paymentClient.ListPayments(status)
}
