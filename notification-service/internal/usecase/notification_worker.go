package usecase

import (
	"fmt"
	"log"
	"time"

	"notification-service/internal/domain"
)

type NotificationWorker struct {
	provider         domain.NotificationProvider
	idempotencyStore domain.IdempotencyStore
}

func NewNotificationWorker(provider domain.NotificationProvider, store domain.IdempotencyStore) *NotificationWorker {
	return &NotificationWorker{
		provider:         provider,
		idempotencyStore: store,
	}
}

func (w *NotificationWorker) ProcessEvent(messageID string, event domain.NotificationEvent) error {

	if w.idempotencyStore != nil {
		processed, err := w.idempotencyStore.IsProcessed(messageID)
		if err != nil {
			log.Printf("Warning: failed to check idempotency for msg %s: %v", messageID, err)
		} else if processed {
			log.Printf("Duplicate message ignored: %s", messageID)
			return nil
		}
	}

	subject := fmt.Sprintf("Payment Update for Order #%s", event.OrderID)
	body := fmt.Sprintf("Your payment of $%v is %s", float64(event.Amount)/100.0, event.Status)

	maxRetries := 3
	backoffDuration := 2 * time.Second

	var err error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		err = w.provider.SendEmail(event.CustomerEmail, subject, body)
		if err == nil {
			break
		}

		log.Printf("Attempt %d failed to send email: %v", attempt, err)
		if attempt < maxRetries {
			log.Printf("Retrying in %v...", backoffDuration)
			time.Sleep(backoffDuration)
			backoffDuration *= 2
		}
	}

	if err != nil {

		return fmt.Errorf("failed to process notification after %d attempts: %v", maxRetries, err)
	}

	if w.idempotencyStore != nil {
		if err := w.idempotencyStore.MarkProcessed(messageID); err != nil {
			log.Printf("Warning: failed to mark message %s as processed in Redis: %v", messageID, err)
		}
	}

	return nil
}
