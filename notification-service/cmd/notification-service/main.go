package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"notification-service/internal/domain"
	"notification-service/internal/infrastructure"
	"notification-service/internal/usecase"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "localhost:6379"
	}
	idempotencyStore, err := infrastructure.NewRedisIdempotencyStore(redisURL)
	if err != nil {
		log.Printf("Warning: Could not connect to Redis for idempotency: %v", err)
	}

	providerMode := os.Getenv("PROVIDER_MODE")
	var provider domain.NotificationProvider
	if providerMode == "REAL" {
		provider = infrastructure.NewRealProvider()
	} else {
		provider = infrastructure.NewSimulatedProvider()
	}

	worker := usecase.NewNotificationWorker(provider, idempotencyStore)

	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}

	var conn *amqp.Connection
	for i := 0; i < 5; i++ {
		conn, err = amqp.Dial(rabbitURL)
		if err == nil {
			break
		}
		log.Printf("Failed to connect to RabbitMQ, retrying in 2 seconds... (%v)", err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ after retries: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	err = ch.ExchangeDeclare("payment.dlx", "direct", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to declare DLX: %v", err)
	}

	_, err = ch.QueueDeclare("payment.dlq", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to declare DLQ: %v", err)
	}

	err = ch.QueueBind("payment.dlq", "dlq_key", "payment.dlx", false, nil)
	if err != nil {
		log.Fatalf("Failed to bind DLQ: %v", err)
	}

	args := amqp.Table{
		"x-dead-letter-exchange":    "payment.dlx",
		"x-dead-letter-routing-key": "dlq_key",
	}

	q, err := ch.QueueDeclare("payment.completed", true, false, false, false, args)
	if err != nil {
		log.Fatalf("Failed to declare queue: %v", err)
	}

	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s (MsgId: %s)", d.Body, d.MessageId)

			var event domain.NotificationEvent
			if err := json.Unmarshal(d.Body, &event); err != nil {
				log.Printf("Error parsing message: %v", err)
				d.Nack(false, false)
				continue
			}

			if event.Amount == 99999 {
				log.Printf("Simulating permanent error for amount 99999, moving to DLQ")
				d.Nack(false, false)
				continue
			}

			if err := worker.ProcessEvent(d.MessageId, event); err != nil {
				log.Printf("Worker failed to process event: %v", err)
				d.Nack(false, false)
				continue
			}

			d.Ack(false)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down gracefully...")
}
