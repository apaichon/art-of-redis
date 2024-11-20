// payment_worker.go
package main

import (
    "context"
    "encoding/json"
    "log"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type PaymentProcessor struct {
    rdb *redis.Client
    ctx context.Context
}

func NewPaymentProcessor() *PaymentProcessor {
    return &PaymentProcessor{
        rdb: redis.NewClient(&redis.Options{
            Addr: "localhost:6379",
        }),
        ctx: context.Background(),
    }
}

func (p *PaymentProcessor) ProcessPayment(payload string) error {
    var paymentEvent map[string]interface{}
    if err := json.Unmarshal([]byte(payload), &paymentEvent); err != nil {
        return err
    }

    // Simulate payment gateway processing
    time.Sleep(2 * time.Second)
    
    // Update payment status
    log.Printf("Processing payment for ticket %v", paymentEvent["ticket_id"])
    
    // Publish success notification
    notification := map[string]interface{}{
        "type":      "payment_success",
        "ticket_id": paymentEvent["ticket_id"],
        "user_id":   paymentEvent["user_id"],
        "timestamp": time.Now(),
    }
    notifJSON, _ := json.Marshal(notification)
    return p.rdb.Publish(p.ctx, "notification_events", string(notifJSON)).Err()
}

func main() {
    processor := NewPaymentProcessor()
    pubsub := processor.rdb.Subscribe(processor.ctx, "payment_events")
    defer pubsub.Close()

    fmt.Println("Payment worker started")
    
    ch := pubsub.Channel()
    for msg := range ch {
        if err := processor.ProcessPayment(msg.Payload); err != nil {
            log.Printf("Error processing payment: %v", err)
        }
    }

	
}