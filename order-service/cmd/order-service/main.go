package main

import (
	"database/sql"
	"log"
	"net"
	"os"
	"time"

	orderv1 "github.com/Bleuble/my-grpc-generated/order/v1"
	"order-service/internal/app"
	"order-service/internal/infrastructure"
	"order-service/internal/repository"
	"order-service/internal/transport/grpc_handler"
	"order-service/internal/transport/http"
	"order-service/internal/usecase"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

func main() {

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:admin@localhost:5433/order_db?sslmode=disable"
	}

	paymentServiceAddr := os.Getenv("PAYMENT_SERVICE_ADDR")
	if paymentServiceAddr == "" {
		paymentServiceAddr = "localhost:50051"
	}

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50052"
	}

	restPort := os.Getenv("REST_PORT")
	if restPort == "" {
		restPort = "8080"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping db: %v", err)
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "localhost:6379"
	}

	orderCache, err := infrastructure.NewRedisOrderCache(redisURL, 5*time.Minute)
	if err != nil {
		log.Printf("Warning: Redis cache not available: %v", err)
	}

	conn, err := grpc.Dial(paymentServiceAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect to payment service: %v", err)
	}
	defer conn.Close()
	paymentClient := app.NewGrpcPaymentClient(conn)

	orderRepo := repository.NewPostgresOrderRepository(db)
	orderUseCase := usecase.NewOrderUseCase(orderRepo, paymentClient, orderCache)

	go func() {
		lis, err := net.Listen("tcp", ":"+grpcPort)
		if err != nil {
			log.Fatalf("failed to listen for gRPC: %v", err)
		}

		s := grpc.NewServer()
		orderv1.RegisterOrderTrackingServiceServer(s, grpc_handler.NewOrderHandler(orderUseCase))

		log.Printf("Order Tracking gRPC Service is running on port %s...", grpcPort)
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve gRPC: %v", err)
		}
	}()

	router := gin.Default()

	if orderCache != nil {

		router.Use(http.RateLimiterMiddleware(orderCache.GetClient(), 10, time.Minute))
	}

	orderHandler := http.NewOrderHandler(orderUseCase)
	orderHandler.RegisterRoutes(router)

	log.Printf("Order HTTP Service is running on port %s...", restPort)
	if err := router.Run(":" + restPort); err != nil {
		log.Fatalf("failed to run REST server: %v", err)
	}
}
