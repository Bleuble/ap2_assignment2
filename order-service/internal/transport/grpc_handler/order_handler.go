package grpc_handler

import (
	"log"

	orderv1 "github.com/Bleuble/my-grpc-generated/order/v1"
	"order-service/internal/usecase"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type OrderHandler struct {
	orderv1.UnimplementedOrderTrackingServiceServer
	useCase *usecase.OrderUseCase
}

func NewOrderHandler(uc *usecase.OrderUseCase) *OrderHandler {
	return &OrderHandler{useCase: uc}
}

func (h *OrderHandler) SubscribeToOrderUpdates(req *orderv1.OrderRequest, stream orderv1.OrderTrackingService_SubscribeToOrderUpdatesServer) error {
	log.Printf("New subscriber for Order: %s", req.OrderId)

	ch, cleanup := h.useCase.Subscribe(req.OrderId)
	defer cleanup()

	order, err := h.useCase.GetOrder(req.OrderId)
	if err == nil && order != nil {
		err := stream.Send(&orderv1.OrderStatusUpdate{
			OrderId:   order.ID,
			Status:    order.Status,
			UpdatedAt: timestamppb.New(order.CreatedAt),
		})
		if err != nil {
			return err
		}
	}

	for {
		select {
		case <-stream.Context().Done():
			log.Printf("Subscriber for Order %s disconnected", req.OrderId)
			return nil
		case status, ok := <-ch:
			if !ok {
				return nil
			}
			err := stream.Send(&orderv1.OrderStatusUpdate{
				OrderId:   req.OrderId,
				Status:    status,
				UpdatedAt: timestamppb.Now(),
			})
			if err != nil {
				log.Printf("Error sending update for Order %s: %v", req.OrderId, err)
				return err
			}
		}
	}
}
