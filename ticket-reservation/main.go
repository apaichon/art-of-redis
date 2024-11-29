// main.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

const (
	ReservationTimeout = 15 * time.Minute
	RedisAddr          = "localhost:6379"
)

type Ticket struct {
	ID         string    `json:"id"`
	Status     string    `json:"status"` // available, reserved, paid
	ReservedAt time.Time `json:"reserved_at,omitempty"`
	UserID     string    `json:"user_id,omitempty"`
}

type WaitingList struct {
	TicketID string    `json:"ticket_id"`
	UserID   string    `json:"user_id"`
	AddedAt  time.Time `json:"added_at"`
}

var (
	rdb *redis.Client
	ctx = context.Background()
)

func init() {
	rdb = redis.NewClient(&redis.Options{
		Addr: RedisAddr,
	})
}

// Core API Handler
func reserveTicketHandler(w http.ResponseWriter, r *http.Request) {
	var ticket Ticket
	if err := json.NewDecoder(r.Body).Decode(&ticket); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if ticket exists and is available
	exists, err := rdb.Get(ctx, "ticket:"+ticket.ID).Result()
	if err != nil || exists == "" {
		http.Error(w, "Ticket not found", http.StatusNotFound)
		return
	}

	fmt.Println( "time: ", time.Now().Format(time.RFC3339), "exists: ", exists)

	var currentTicket Ticket
	json.Unmarshal([]byte(exists), &currentTicket)


	if currentTicket.Status == "paid" {
		http.Error(w, "Ticket already paid", http.StatusBadRequest)
		return
	} else if currentTicket.Status == "reserved" {
		// Add to waiting list
		waitingItem := WaitingList{
			TicketID: ticket.ID,
			UserID:   ticket.UserID,
			AddedAt:  time.Now(),
		}
		itemJSON, _ := json.Marshal(waitingItem)
		rdb.RPush(ctx, "waitinglist:"+waitingItem.TicketID, string(itemJSON)).Err()
		
		// New response message
		w.WriteHeader(http.StatusAccepted) // Indicate that the user has been added to the waiting list
		response := map[string]string{
			"message": "You have been added to the waiting list for ticket " + ticket.ID,
		}
		json.NewEncoder(w).Encode(response) // Send the response message
		return
	}

	// Try to reserve the ticket
	err = rdb.Watch(ctx, func(tx *redis.Tx) error {
		ticket.Status = "reserved"
		ticket.ReservedAt = time.Now()

		ticketJSON, _ := json.Marshal(ticket)
		tx.Set(ctx, "ticket:"+ticket.ID, string(ticketJSON), ReservationTimeout)

		return nil
	}, "ticket:"+ticket.ID)

	if err != nil {
		http.Error(w, "Failed to reserve ticket", http.StatusConflict)
		return
	}

	// Set expiration callback
	rdb.Set(ctx, "expire:"+ticket.ID, ticket.UserID, ReservationTimeout)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ticket)
}

func addToWaitingListHandler(w http.ResponseWriter, r *http.Request) {
	var waitingItem WaitingList
	if err := json.NewDecoder(r.Body).Decode(&waitingItem); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	waitingItem.AddedAt = time.Now()
	itemJSON, _ := json.Marshal(waitingItem)

	err := rdb.RPush(ctx, "waitinglist:"+waitingItem.TicketID, string(itemJSON)).Err()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func payTicketHandler(w http.ResponseWriter, r *http.Request) {
	var ticket Ticket
	if err := json.NewDecoder(r.Body).Decode(&ticket); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Start Redis Transaction
	txn := rdb.TxPipeline()

	// Check if ticket is reserved by the user
	ticketJSON, err := rdb.Get(ctx, "ticket:"+ticket.ID).Result()
	if err != nil {
		http.Error(w, "Ticket not found", http.StatusNotFound)
		return
	}

	var currentTicket Ticket
	json.Unmarshal([]byte(ticketJSON), &currentTicket)

	if currentTicket.Status != "reserved" || currentTicket.UserID != ticket.UserID {
		http.Error(w, "Invalid ticket status or ownership", http.StatusBadRequest)
		return
	}

	// Update ticket status
	currentTicket.Status = "paid"
	updatedTicketJSON, _ := json.Marshal(currentTicket)

	txn.Set(ctx, "ticket:"+ticket.ID, string(updatedTicketJSON), 0)
	txn.Del(ctx, "expire:"+ticket.ID)

	// Execute transaction
	if _, err := txn.Exec(ctx); err != nil {
		http.Error(w, "Failed to process payment", http.StatusInternalServerError)
		return
	}

	// Publish payment event
	paymentEvent := map[string]interface{}{
		"ticket_id": ticket.ID,
		"user_id":   ticket.UserID,
		"timestamp": time.Now(),
	}
	eventJSON, _ := json.Marshal(paymentEvent)
	rdb.Publish(ctx, "payment_events", string(eventJSON))

	w.WriteHeader(http.StatusOK)
}

// New API Handler for creating a ticket
func createTicketHandler(w http.ResponseWriter, r *http.Request) {
	var ticket Ticket
	if err := json.NewDecoder(r.Body).Decode(&ticket); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set initial status and reserved time
	ticket.Status = "available"
	ticket.ReservedAt = time.Time{} // Reset reserved time
	ticketJSON, _ := json.Marshal(ticket)

	// Store the ticket in Redis
	err := rdb.Set(ctx, "ticket:"+ticket.ID, string(ticketJSON), 0).Err()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ticket)
}

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/tickets/reserve", reserveTicketHandler).Methods("POST")
	router.HandleFunc("/tickets/waiting-list", addToWaitingListHandler).Methods("POST")
	router.HandleFunc("/tickets/pay", payTicketHandler).Methods("POST")
	router.HandleFunc("/tickets/create", createTicketHandler).Methods("POST")

	fmt.Println("Server started on port 9005")
	log.Fatal(http.ListenAndServe(":9005", router))

}

/*
API:
1. Reserve Ticket
2. Add to Waiting List
3. Pay Ticket
4. Create Ticket

Payment Worker:
1. Listen to payment_events
2. Update ticket status to paid
3. Remove from waiting list

Notifications:
1. Payment Worker: When a payment is successful, it publishes a notification to the "ticket_paid" channel
2. Core API: Listens to the "ticket_paid" channel and updates the ticket status to "paid"
3. Core API: Listens to the "ticket_reserved" channel and updates the ticket status to "reserved"

*/
