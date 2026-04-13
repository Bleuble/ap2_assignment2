package domain

type PaymentClient interface {
	AuthorizePayment(orderID string, amount int64) (string, error)
}
