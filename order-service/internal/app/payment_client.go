package app

import (
	"bytes"
	"context" // Added for gRPC
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	pb "github.com/Bleuble/my-grpc-generated/payment/v1" // Added generated proto
	"google.golang.org/grpc"                           // Added gRPC
)

type HttpPaymentClient struct {
	client  *http.Client
	baseURL string
}

func NewHttpPaymentClient(baseURL string) *HttpPaymentClient {
	return &HttpPaymentClient{
		client: &http.Client{
			Timeout: 2 * time.Second,
		},
		baseURL: baseURL,
	}
}

type paymentPayload struct {
	OrderID string `json:"order_id"`
	Amount  int64  `json:"amount"`
}

type paymentResponse struct {
	Status        string `json:"status"`
	TransactionID string `json:"transaction_id"`
}

func (c *HttpPaymentClient) AuthorizePayment(orderID string, amount int64) (string, error) {
	url := fmt.Sprintf("%s/payments", c.baseURL)

	payload := paymentPayload{
		OrderID: orderID,
		Amount:  amount,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("payment rejected with status %d", resp.StatusCode)
	}

	var res paymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", err
	}

	if res.Status != "Authorized" {
		return "", fmt.Errorf("payment declined")
	}

	return res.TransactionID, nil
}

// GrpcPaymentClient implements domain.PaymentClient using gRPC
type GrpcPaymentClient struct {
	client pb.PaymentServiceClient
}

func NewGrpcPaymentClient(conn *grpc.ClientConn) *GrpcPaymentClient {
	return &GrpcPaymentClient{
		client: pb.NewPaymentServiceClient(conn),
	}
}

func (c *GrpcPaymentClient) AuthorizePayment(orderID string, amount int64) (string, error) {
	// Call the remote Payment Service via gRPC
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resp, err := c.client.ProcessPayment(ctx, &pb.PaymentRequest{
		OrderId:  orderID,
		Amount:   float64(amount),
		Currency: "USD",
	})
	if err != nil {
		return "", fmt.Errorf("gRPC payment failed: %v", err)
	}

	if resp.Status != "SUCCESS" {
		return "", fmt.Errorf("payment declined: %s", resp.Status)
	}

	return resp.TransactionId, nil
}
