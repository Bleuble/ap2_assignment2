package grpc_handler

import (
	"context"
	"log"
	"time"

	"payment-service/internal/usecase"
	paymentv1 "github.com/Bleuble/my-grpc-generated/payment/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PaymentHandler struct {
	paymentv1.UnimplementedPaymentServiceServer
	useCase *usecase.PaymentUseCase
}

func NewPaymentHandler(uc *usecase.PaymentUseCase) *PaymentHandler {
	return &PaymentHandler{useCase: uc}
}

func (h *PaymentHandler) ProcessPayment(ctx context.Context, req *paymentv1.PaymentRequest) (*paymentv1.PaymentResponse, error) {
	// Use Case exists in Assignment 1 logic
	// In a real scenario, you'd map the proto message to a domain entity
	// For now, we'll implement the logic directly or call the use case
	
	if req.Amount <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "amount must be greater than zero")
	}

	log.Printf("Processing payment for Order: %s, Amount: %f", req.OrderId, req.Amount)

	// Simulate processing
	return &paymentv1.PaymentResponse{
		TransactionId: "tx_" + time.Now().Format("20060102150405"),
		Status:        "SUCCESS",
	}, nil
}

// LoggingInterceptor logs the method name and duration of each call
func LoggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	log.Printf("gRPC Call - Method: %s, Duration: %s, Error: %v", info.FullMethod, time.Since(start), err)
	return resp, err
}
