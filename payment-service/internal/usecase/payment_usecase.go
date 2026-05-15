package usecase

import (
	"fmt"
	"payment-service/internal/domain"
)

type PaymentUseCase struct {
	repo      domain.PaymentRepository
	publisher domain.EventPublisher
}

func NewPaymentUseCase(repo domain.PaymentRepository, publisher domain.EventPublisher) *PaymentUseCase {
	return &PaymentUseCase{repo: repo, publisher: publisher}
}

func (uc *PaymentUseCase) ProcessPayment(orderID string, amount int64) (*domain.Payment, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("amount must be greater than zero")
	}

	payment := domain.ProcessPayment(orderID, amount)

	if err := uc.repo.Save(payment); err != nil {
		return nil, fmt.Errorf("failed to save payment: %v", err)
	}

	if uc.publisher != nil && payment.Status == "Authorized" {
		event := domain.PaymentEvent{
			OrderID:       payment.OrderID,
			Amount:        payment.Amount,
			CustomerEmail: "user_" + payment.OrderID + "@example.com",
			Status:        payment.Status,
		}
		if err := uc.publisher.PublishPaymentCompleted(event); err != nil {
			fmt.Printf("Warning: failed to publish payment event: %v\n", err)
		}
	}

	return payment, nil
}

func (uc *PaymentUseCase) GetPaymentStatus(orderID string) (*domain.Payment, error) {
	return uc.repo.GetByOrderID(orderID)
}

func (uc *PaymentUseCase) ListPayments(status string) ([]*domain.Payment, error) {
	return uc.repo.ListByStatus(status)
}
