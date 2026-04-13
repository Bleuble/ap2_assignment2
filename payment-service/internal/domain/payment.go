package domain

import "github.com/google/uuid"

type Payment struct {
	ID            string
	OrderID       string
	TransactionID string
	Amount        int64
	Status        string
}

func ProcessPayment(orderID string, amount int64) *Payment {
	p := &Payment{
		ID:            uuid.New().String(),
		OrderID:       orderID,
		TransactionID: uuid.New().String(),
		Amount:        amount,
	}

	if amount > 100000 {
		p.Status = "Declined"
	} else {
		p.Status = "Authorized"
	}

	return p
}
