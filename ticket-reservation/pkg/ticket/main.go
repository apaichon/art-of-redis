package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

type TicketSystem struct {
	redisClient *redis.Client
}

func NewTicketSystem() *TicketSystem {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // Update if Redis is on another host
	})
	return &TicketSystem{redisClient: client}
}

// Reserve a ticket
func (ts *TicketSystem) Reserve(ticketID, userID string) error {
	key := fmt.Sprintf("ticket:%s", ticketID)
	waitingListKey := fmt.Sprintf("waiting:%s", ticketID)

	return ts.redisClient.Watch(ctx, func(tx *redis.Tx) error {
		available, err := tx.Get(ctx, key).Int()
		if err == redis.Nil || available <= 0 {
			// Add to waiting list if no tickets available
			return tx.LPush(ctx, waitingListKey, userID).Err()
		} else if err != nil {
			return err
		}

		// Reserve ticket
		pipe := tx.Pipeline()
		pipe.Decr(ctx, key)
		pipe.HSet(ctx, fmt.Sprintf("reservation:%s", userID), "ticketID", ticketID, "status", "reserved")
		_, err = pipe.Exec(ctx)
		return err
	}, key)
}

// Cancel a reservation
func (ts *TicketSystem) Cancel(userID string) error {
	reservationKey := fmt.Sprintf("reservation:%s", userID)
	reservation, err := ts.redisClient.HGetAll(ctx, reservationKey).Result()
	if err != nil || len(reservation) == 0 {
		return fmt.Errorf("no reservation found for user %s", userID)
	}

	ticketID := reservation["ticketID"]
	key := fmt.Sprintf("ticket:%s", ticketID)
	waitingListKey := fmt.Sprintf("waiting:%s", ticketID)

	return ts.redisClient.Watch(ctx, func(tx *redis.Tx) error {
		pipe := tx.Pipeline()
		pipe.Del(ctx, reservationKey)
		pipe.Incr(ctx, key)

		// Notify next user in waiting list
		nextUser, err := tx.LPop(ctx, waitingListKey).Result()
		if err == nil {
			pipe.HSet(ctx, fmt.Sprintf("reservation:%s", nextUser), "ticketID", ticketID, "status", "reserved")
			pipe.Decr(ctx, key)
		}
		_, err = pipe.Exec(ctx)
		return err
	}, reservationKey, key, waitingListKey)
}

// Make a payment
func (ts *TicketSystem) Pay(userID string) error {
	reservationKey := fmt.Sprintf("reservation:%s", userID)
	reservation, err := ts.redisClient.HGetAll(ctx, reservationKey).Result()
	if err != nil || len(reservation) == 0 {
		return fmt.Errorf("no reservation found for user %s", userID)
	}

	status := reservation["status"]
	if status != "reserved" {
		return fmt.Errorf("cannot pay for reservation with status %s", status)
	}

	pipe := ts.redisClient.Pipeline()
	pipe.HSet(ctx, reservationKey, "status", "paid")
	pipe.Publish(ctx, "notifications", fmt.Sprintf("User %s paid for ticket %s", userID, reservation["ticketID"]))
	_, err = pipe.Exec(ctx)
	return err
}

// Notify users
func (ts *TicketSystem) Notify(channel string) {
	pubsub := ts.redisClient.Subscribe(ctx, channel)
	defer pubsub.Close()

	for msg := range pubsub.Channel() {
		log.Printf("Notification: %s", msg.Payload)
	}
}

func main() {
	ts := NewTicketSystem()
	defer ts.redisClient.Close()

	// Simulate ticket availability
	ticketID := "123"
	ts.redisClient.Set(ctx, fmt.Sprintf("ticket:%s", ticketID), 5, 0)

	// List multiple tickets
	ticketIDs := []string{ "456", "789", "101112"}
	for _, id := range ticketIDs {
		ts.redisClient.Set(ctx, fmt.Sprintf("ticket:%s", id), 5, 0)
	}

	// Generate random user IDs and reserve tickets
	go func() {
		for i := 1; i <= 10; i++ {
			userID := fmt.Sprintf("user%d", i)
			for _, id := range ticketIDs {
				if err := ts.Reserve(id, userID); err != nil {
					log.Printf("Reserve error for %s on ticket %s: %v", userID, id, err)
				}
			}
		}
	}()

	// Cancel reservation
	go func() {
		time.Sleep(2 * time.Second)
		if err := ts.Cancel("user1"); err != nil {
			log.Printf("Cancel error: %v", err)
		}
	}()

	// Make payment
	go func() {
		time.Sleep(1 * time.Second)
		if err := ts.Pay("user2"); err != nil {
			log.Printf("Pay error: %v", err)
		}
	}()

	// Listen for notifications
	go ts.Notify("notifications")

	// Prevent main from exiting immediately
	time.Sleep(5 * time.Second)
}
