package domain

import (
	"fmt"
	"time"
)

type Order struct {
	ID             string
	CustomerID     string
	ItemName       string
	Amount         int64
	Status         string
	IdempotencyKey string
	CreatedAt      time.Time
}

func (o *Order) Validate() error {
	if o.CustomerID == "" {
		return fmt.Errorf("customer_id is required")
	}
	if o.ItemName == "" {
		return fmt.Errorf("item_name is required")
	}
	if o.Amount <= 0 {
		return fmt.Errorf("amount must be greater than zero")
	}
	return nil
}

func (o *Order) DanaPaid() {
	o.Status = "Paid"
}

func (o *Order) DanaFailed() {
	o.Status = "Failed"
}

func (o *Order) Cancel() error {
	if o.Status == "Paid" {
		return fmt.Errorf("paid orders cannot be cancelled")
	}
	if o.Status == "Cancelled" {
		return fmt.Errorf("order is already cancelled")
	}
	o.Status = "Cancelled"
	return nil
}
