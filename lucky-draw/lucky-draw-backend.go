// main.go
package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type Server struct {
	redis    *redis.Client
	ctx      context.Context
	upgrader websocket.Upgrader
	clients  map[*websocket.Conn]bool
}

type Draw struct {
	ID            string    `json:"id"`
	Number        string    `json:"number"`
	Status        string    `json:"status"` // pending, spinning, completed
	WinningNumber string    `json:"winningNumber,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
	CompletedAt   time.Time `json:"completedAt,omitempty"`
}

type Winner struct {
	DrawID      string    `json:"drawId"`
	Number      string    `json:"number"`
	Prize       string    `json:"prize"`
	ClaimedAt   time.Time `json:"claimedAt"`
}

func NewServer() *Server {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})

	return &Server{
		redis: rdb,
		ctx:   context.Background(),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		clients: make(map[*websocket.Conn]bool),
	}
}

// Generate a random 6-digit number
func generateNumber() string {
	max := big.NewInt(1000000)
	n, _ := rand.Int(rand.Reader, max)
	return fmt.Sprintf("%06d", n.Int64())
}

// Start a new draw
func (s *Server) startDraw(w http.ResponseWriter, r *http.Request) {
	draw := Draw{
		ID:        fmt.Sprintf("draw:%d", time.Now().UnixNano()),
		Number:    generateNumber(),
		Status:    "spinning",
		CreatedAt: time.Now(),
	}

	// Store draw in Redis
	drawJSON, _ := json.Marshal(draw)
	err := s.redis.Set(s.ctx, draw.ID, drawJSON, 24*time.Hour).Err()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Broadcast draw start
	s.broadcastDrawUpdate(draw)

	// Simulate wheel spinning
	go func() {
		time.Sleep(5 * time.Second)
		draw.Status = "completed"
		draw.WinningNumber = generateNumber()
		draw.CompletedAt = time.Now()

		// Update draw in Redis
		drawJSON, _ := json.Marshal(draw)
		s.redis.Set(s.ctx, draw.ID, drawJSON, 24*time.Hour)

		// Store winning number
		s.redis.SAdd(s.ctx, "winning_numbers", draw.WinningNumber)

		// Broadcast draw completion
		s.broadcastDrawUpdate(draw)
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(draw)
}

// Claim a prize
func (s *Server) claimPrize(w http.ResponseWriter, r *http.Request) {
	var claim struct {
		DrawID string `json:"drawId"`
		Number string `json:"number"`
	}

	if err := json.NewDecoder(r.Body).Decode(&claim); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get draw details
	drawJSON, err := s.redis.Get(s.ctx, claim.DrawID).Result()
	if err != nil {
		http.Error(w, "Draw not found", http.StatusNotFound)
		return
	}

	var draw Draw
	json.Unmarshal([]byte(drawJSON), &draw)

	// Verify winning number
	if draw.WinningNumber != claim.Number {
		http.Error(w, "Number does not match winning number", http.StatusBadRequest)
		return
	}

	// Check if already claimed
	claimed, err := s.redis.SIsMember(s.ctx, "claimed_prizes", claim.Number).Result()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if claimed {
		http.Error(w, "Prize already claimed", http.StatusConflict)
		return
	}

	// Record prize claim
	winner := Winner{
		DrawID:    claim.DrawID,
		Number:    claim.Number,
		Prize:     "Prize Package A", // In real app, determine prize based on rules
		ClaimedAt: time.Now(),
	}

	winnerJSON, _ := json.Marshal(winner)
	pipe := s.redis.Pipeline()
	pipe.Set(s.ctx, fmt.Sprintf("winner:%s", claim.Number), winnerJSON, 30*24*time.Hour)
	pipe.SAdd(s.ctx, "claimed_prizes", claim.Number)
	_, err = pipe.Exec(s.ctx)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Broadcast winner
	s.broadcastWinner(winner)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(winner)
}

// WebSocket handling
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	s.clients[conn] = true
	defer delete(s.clients, conn)

	// Keep connection alive and handle incoming messages
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

// Broadcast updates
func (s *Server) broadcastDrawUpdate(draw Draw) {
	message := map[string]interface{}{
		"type": "draw_update",
		"data": draw,
	}
	s.broadcast(message)
}

func (s *Server) broadcastWinner(winner Winner) {
	message := map[string]interface{}{
		"type": "winner_announcement",
		"data": winner,
	}
	s.broadcast(message)
}

func (s *Server) broadcast(message interface{}) {
	messageJSON, _ := json.Marshal(message)
	for client := range s.clients {
		err := client.WriteMessage(websocket.TextMessage, messageJSON)
		if err != nil {
			log.Printf("Error sending message: %v", err)
			client.Close()
			delete(s.clients, client)
		}
	}
}

func main() {
	server := NewServer()
	router := mux.NewRouter()

	router.HandleFunc("/api/draw/start", server.startDraw).Methods("POST")
	router.HandleFunc("/api/draw/claim", server.claimPrize).Methods("POST")
	router.HandleFunc("/ws", server.handleWebSocket)

	log.Fatal(http.ListenAndServe(":8080", router))
}
