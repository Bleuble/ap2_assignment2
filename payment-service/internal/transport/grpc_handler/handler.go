package grpc_handler

import (
	"context"
	"log"
	"time"

	paymentv1 "github.com/Bleuble/my-grpc-generated/payment/v1"
	"payment-service/internal/usecase"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PaymentHandler struct {
	paymentv1.UnimplementedPaymentServiceServer
	useCase *usecase.PaymentUseCase
}

func NewPaymentHandler(uc *usecase.PaymentUseCase) *PaymentHandler {
	return &PaymentHandler{useCase: uc}
}

func (h *PaymentHandler) ProcessPayment(ctx context.Context, req *paymentv1.PaymentRequest) (*paymentv1.PaymentResponse, error) {
	log.Printf("Processing payment for Order: %s, Amount: %f", req.OrderId, req.Amount)

	payment, err := h.useCase.ProcessPayment(req.OrderId, int64(req.Amount))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to process payment: %v", err)
	}

	return &paymentv1.PaymentResponse{
		TransactionId: payment.TransactionID,
		Status:        payment.Status,
		ProcessedAt:   timestamppb.Now(),
	}, nil
}

// LoggingInterceptor logs the method name and duration of each call
func LoggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	log.Printf("gRPC Call - Method: %s, Duration: %s, Error: %v", info.FullMethod, time.Since(start), err)
	return resp, err
}
