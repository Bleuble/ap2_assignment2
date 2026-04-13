package usecase

import (
	"fmt"
	"payment-service/internal/domain"
)

type PaymentUseCase struct {
	repo domain.PaymentRepository
}

func NewPaymentUseCase(repo domain.PaymentRepository) *PaymentUseCase {
	return &PaymentUseCase{repo: repo}
}

func (uc *PaymentUseCase) ProcessPayment(orderID string, amount int64) (*domain.Payment, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("amount must be greater than zero")
	}

	payment := domain.ProcessPayment(orderID, amount)

	if err := uc.repo.Save(payment); err != nil {
		return nil, fmt.Errorf("failed to save payment: %v", err)
	}

	return payment, nil
}

func (uc *PaymentUseCase) GetPaymentStatus(orderID string) (*domain.Payment, error) {
	return uc.repo.GetByOrderID(orderID)
}
