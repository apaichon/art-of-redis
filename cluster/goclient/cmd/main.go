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

// TicketData represents a ticket in the system
type TicketData struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	UserID    string    `json:"userId"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ClusterClient wraps Redis cluster operations
type ClusterClient struct {
	client *redis.ClusterClient
}

// NewClusterClient creates a new Redis cluster client
func NewClusterClient() (*ClusterClient, error) {
	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: []string{
			"redis1:7001",
			"redis2:7002",
			"redis3:7003",
		},
		// Routing settings
		RouteByLatency: true,
		RouteRandomly:  true,

		// Connection pool settings
		PoolSize:     10,
		MinIdleConns: 5,

		// Timeouts
		ReadTimeout:  time.Second * 5,
		WriteTimeout: time.Second * 5,
		PoolTimeout:  time.Second * 4,
	})

	// Test connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis cluster: %v", err)
	}

	return &ClusterClient{client: client}, nil
}

// Close closes the cluster client
func (c *ClusterClient) Close() error {
	return c.client.Close()
}

// CreateTicket creates a new ticket
func (c *ClusterClient) CreateTicket(ctx context.Context, ticket *TicketData) error {
	ticket.CreatedAt = time.Now()
	ticket.UpdatedAt = time.Now()

	data, err := json.Marshal(ticket)
	if err != nil {
		return fmt.Errorf("failed to marshal ticket: %v", err)
	}

	key := fmt.Sprintf("ticket:{%s}", ticket.ID) // Ensures same hash slot
	err = c.client.Set(ctx, key, data, 0).Err()
	if err != nil {
		return fmt.Errorf("failed to create ticket: %v", err)
	}

	return nil
}

// GetTicket retrieves a ticket by ID
func (c *ClusterClient) GetTicket(ctx context.Context, ticketID string) (*TicketData, error) {
	key := fmt.Sprintf("ticket:{%s}", ticketID)
	data, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("ticket not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get ticket: %v", err)
	}

	var ticket TicketData
	if err := json.Unmarshal([]byte(data), &ticket); err != nil {
		return nil, fmt.Errorf("failed to unmarshal ticket: %v", err)
	}

	return &ticket, nil
}

// UpdateTicketStatus updates ticket status
func (c *ClusterClient) UpdateTicketStatus(ctx context.Context, ticketID, status string) error {
	ticket, err := c.GetTicket(ctx, ticketID)
	if err != nil {
		return err
	}

	ticket.Status = status
	ticket.UpdatedAt = time.Now()

	return c.CreateTicket(ctx, ticket) // Overwrites existing ticket
}

// DeleteTicket deletes a ticket
func (c *ClusterClient) DeleteTicket(ctx context.Context, ticketID string) error {
	key := fmt.Sprintf("ticket:{%s}", ticketID)
	err := c.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete ticket: %v", err)
	}
	return nil
}

// CreateTicketHandler handles the creation of a new ticket
func (c *ClusterClient) CreateTicketHandler(w http.ResponseWriter, r *http.Request) {
	var ticket TicketData
	if err := json.NewDecoder(r.Body).Decode(&ticket); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := c.CreateTicket(r.Context(), &ticket); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ticket)
}

// GetTicketHandler retrieves a ticket by ID
func (c *ClusterClient) GetTicketHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ticketID := vars["id"]

	ticket, err := c.GetTicket(r.Context(), ticketID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(ticket)
}

// UpdateTicketStatusHandler updates the status of a ticket
func (c *ClusterClient) UpdateTicketStatusHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ticketID := vars["id"]

	var update struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := c.UpdateTicketStatus(r.Context(), ticketID, update.Status); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteTicketHandler deletes a ticket by ID
func (c *ClusterClient) DeleteTicketHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ticketID := vars["id"]

	if err := c.DeleteTicket(r.Context(), ticketID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// StartAPIServer starts the HTTP server
func StartAPIServer(client *ClusterClient) {
	r := mux.NewRouter()
	r.HandleFunc("/tickets", client.CreateTicketHandler).Methods("POST")
	r.HandleFunc("/tickets/{id}", client.GetTicketHandler).Methods("GET")
	r.HandleFunc("/tickets/{id}/status", client.UpdateTicketStatusHandler).Methods("PUT")
	r.HandleFunc("/tickets/{id}", client.DeleteTicketHandler).Methods("DELETE")

	http.ListenAndServe(":8080", r) // Start the server on port 8080
}

// Example usage
func main() {
	client, err := NewClusterClient()
	if err != nil {
		log.Fatalf("Failed to create cluster client: %v", err)
	}
	defer client.Close()

	go StartAPIServer(client) // Start the API server in a goroutine

	ctx := context.Background()

	// Create a ticket
	ticket := &TicketData{
		ID:     "123",
		Status: "new",
		UserID: "user456",
	}

	if err := client.CreateTicket(ctx, ticket); err != nil {
		log.Printf("Failed to create ticket: %v", err)
		return
	}

	// Get the ticket
	retrievedTicket, err := client.GetTicket(ctx, ticket.ID)
	if err != nil {
		log.Printf("Failed to get ticket: %v", err)
		return
	}
	log.Printf("Retrieved ticket: %+v", retrievedTicket)

	// Update ticket status
	if err := client.UpdateTicketStatus(ctx, ticket.ID, "paid"); err != nil {
		log.Printf("Failed to update ticket: %v", err)
		return
	}

	// Using cluster pipeline for batch operations
	pipe := client.client.Pipeline()

	// Queue multiple operations
	pipe.Set(ctx, "ticket:{123}:payment", "processed", 0)
	pipe.Set(ctx, "ticket:{123}:notification", "sent", 0)

	// Execute pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		log.Printf("Pipeline failed: %v", err)
	}
}

// Optional: Example of cluster transaction (note: limited to keys in same slot)
func (c *ClusterClient) ProcessTicketPayment(ctx context.Context, ticketID string) error {
	key := fmt.Sprintf("ticket:{%s}", ticketID)
	paymentKey := fmt.Sprintf("ticket:{%s}:payment", ticketID)

	txf := func(tx *redis.Tx) error {
		// Get current ticket
		ticket, err := tx.Get(ctx, key).Result()
		if err != nil {
			return err
		}

		// Update ticket and payment status
		pipe := tx.Pipeline()
		pipe.Set(ctx, key, ticket, 0)
		pipe.Set(ctx, paymentKey, "processed", 0)

		_, err = pipe.Exec(ctx)
		return err
	}

	// Retry if the key has been changed
	for i := 0; i < 3; i++ {
		err := c.client.Watch(ctx, txf, key)
		if err == nil {
			return nil
		}
		if err == redis.TxFailedErr {
			continue
		}
		return err
	}

	return fmt.Errorf("transaction failed after 3 retries")
}
