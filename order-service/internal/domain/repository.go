package domain

type OrderRepository interface {
	Create(order *Order) error
	GetByID(id string) (*Order, error)
	GetByIdempotencyKey(key string) (*Order, error)
	UpdateStatus(id string, status string) error
	GetByAmountRange(min, max int64) ([]*Order, error)
}
