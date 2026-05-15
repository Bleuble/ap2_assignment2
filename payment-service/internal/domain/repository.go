package domain

type PaymentRepository interface {
	Save(payment *Payment) error
	GetByOrderID(orderID string) (*Payment, error)
	ListByStatus(status string) ([]*Payment, error)
}
