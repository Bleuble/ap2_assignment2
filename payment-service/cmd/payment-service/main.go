package main

import (
	"database/sql"
	"log"
	"net"
	"os"

	paymentv1 "github.com/Bleuble/my-grpc-generated/payment/v1"
	"payment-service/internal/infrastructure"
	"payment-service/internal/repository"
	"payment-service/internal/transport/grpc_handler"
	"payment-service/internal/transport/http"
	"payment-service/internal/usecase"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

func main() {

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:admin@localhost:5433/payment_db?sslmode=disable"
	}

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}

	restPort := os.Getenv("REST_PORT")
	if restPort == "" {
		restPort = "8081"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping db: %v", err)
	}

	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}
	publisher, err := infrastructure.NewRabbitMQPublisher(rabbitURL)
	if err != nil {
		log.Printf("Warning: Could not connect to RabbitMQ: %v", err)
	} else {
		defer publisher.Close()
	}

	paymentRepo := repository.NewPostgresPaymentRepository(db)
	paymentUseCase := usecase.NewPaymentUseCase(paymentRepo, publisher)

	go func() {
		lis, err := net.Listen("tcp", ":"+grpcPort)
		if err != nil {
			log.Fatalf("failed to listen for gRPC: %v", err)
		}

		s := grpc.NewServer(
			grpc.UnaryInterceptor(grpc_handler.LoggingInterceptor),
		)

		paymentv1.RegisterPaymentServiceServer(s, grpc_handler.NewPaymentHandler(paymentUseCase))

		log.Printf("Payment gRPC Service is running on port %s...", grpcPort)
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve gRPC: %v", err)
		}
	}()

	router := gin.Default()
	paymentHandler := http.NewPaymentHandler(paymentUseCase)
	paymentHandler.RegisterRoutes(router)

	log.Printf("Payment HTTP Service is running on port %s...", restPort)
	if err := router.Run(":" + restPort); err != nil {
		log.Fatalf("failed to run REST server: %v", err)
	}
}
