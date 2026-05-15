package domain

type NotificationEvent struct {
	OrderID       string `json:"order_id"`
	Amount        int64  `json:"amount"`
	CustomerEmail string `json:"customer_email"`
	Status        string `json:"status"`
}

type NotificationProvider interface {
	SendEmail(to string, subject string, body string) error
}

type IdempotencyStore interface {
	IsProcessed(messageID string) (bool, error)
	MarkProcessed(messageID string) error
}
