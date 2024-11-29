// notification_worker.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"
)

type NotificationWorker struct {
	rdb *redis.Client
	ctx context.Context
}

func NewNotificationWorker() *NotificationWorker {
	return &NotificationWorker{
		rdb: redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
		}),
		ctx: context.Background(),
	}
}

func (w *NotificationWorker) ProcessExpiredReservation(ticketID string) error {
	// Get waiting list for the ticket
	waitingListKey := "waitinglist:" + ticketID
	waitingList, err := w.rdb.LRange(w.ctx, waitingListKey, 0, -1).Result()
	if err != nil {
		return err
	}

	// Notify users in waiting list
	for _, item := range waitingList {
		var waitingItem map[string]interface{}
		json.Unmarshal([]byte(item), &waitingItem)

		notification := map[string]interface{}{
			"type":      "ticket_available",
			"ticket_id": ticketID,
			"user_id":   waitingItem["user_id"],
			"message":   "Ticket is now available for reservation",
		}

		// In a real system, this would send to a notification service
		log.Printf("Sending notification: %+v", notification)
	}

	// Clear waiting list
	return w.rdb.Del(w.ctx, waitingListKey).Err()
}

func (w *NotificationWorker) ProcessNotification(payload string) error {
	var notification map[string]interface{}
	if err := json.Unmarshal([]byte(payload), &notification); err != nil {
		return err
	}

	log.Printf("Processing notification: %+v", notification)

	// Handle different notification types
	switch notification["type"] {
	case "reservation_expired":
		return w.ProcessExpiredReservation(notification["ticket_id"].(string))
	case "payment_success":
		// Handle payment success notification
		log.Printf("Payment successful for ticket %v", notification["ticket_id"])
	}

	return nil
}

func main() {
	worker := NewNotificationWorker()
	pubsub := worker.rdb.Subscribe(worker.ctx, "notification_events")
	defer pubsub.Close()

	fmt.Println("Notification worker started")	

	ch := pubsub.Channel()
	for msg := range ch {
		if err := worker.ProcessNotification(msg.Payload); err != nil {
			log.Printf("Error processing notification: %v", err)
		}
	}

}
