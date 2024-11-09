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
	"github.com/gorilla/websocket"
)

type Server struct {
	redis    *redis.Client
	ctx      context.Context
	upgrader websocket.Upgrader
	clients  map[*websocket.Conn]bool
}

type Room struct {
	ID          string    `json:"id"`
	Number      string    `json:"number"`
	Type        string    `json:"type"`
	Price       float64   `json:"price"`
	Available   bool      `json:"available"`
	LastUpdated time.Time `json:"lastUpdated"`
}

type Booking struct {
	ID        string    `json:"id"`
	RoomID    string    `json:"roomId"`
	UserID    string    `json:"userId"`
	CheckIn   time.Time `json:"checkIn"`
	CheckOut  time.Time `json:"checkOut"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

type BookingRequest struct {
	RoomID    string `json:"roomId"`
	UserID    string `json:"userId"`
	CheckIn   string `json:"checkIn"`
	CheckOut  string `json:"checkOut"`
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

// Distributed Lock Implementation
type DistributedLock struct {
	redis     *redis.Client
	ctx       context.Context
	key       string
	value     string
	expiryTTL time.Duration
}

func NewDistributedLock(redis *redis.Client, key string, ttl time.Duration) *DistributedLock {
	return &DistributedLock{
		redis:     redis,
		ctx:      context.Background(),
		key:      fmt.Sprintf("lock:%s", key),
		value:    fmt.Sprintf("lock:%d", time.Now().UnixNano()),
		expiryTTL: ttl,
	}
}

func (dl *DistributedLock) Acquire() (bool, error) {
	return dl.redis.SetNX(dl.ctx, dl.key, dl.value, dl.expiryTTL).Result()
}

func (dl *DistributedLock) Release() error {
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`
	
	result := dl.redis.Eval(dl.ctx, script, []string{dl.key}, dl.value)
	return result.Err()
}

// Room Management
func (s *Server) createRoom(room Room) error {
	room.LastUpdated = time.Now()
	roomJSON, err := json.Marshal(room)
	if err != nil {
		return err
	}

	pipe := s.redis.Pipeline()
	pipe.HSet(s.ctx, "rooms", room.ID, roomJSON)
	pipe.SAdd(s.ctx, "available_rooms", room.ID)
	_, err = pipe.Exec(s.ctx)
	
	if err == nil {
		s.publishRoomUpdate(room)
	}
	
	return err
}

func (s *Server) getRoom(roomID string) (*Room, error) {
	roomJSON, err := s.redis.HGet(s.ctx, "rooms", roomID).Result()
	if err != nil {
		return nil, err
	}

	var room Room
	err = json.Unmarshal([]byte(roomJSON), &room)
	return &room, err
}

func (s *Server) updateRoomAvailability(roomID string, available bool) error {
	room, err := s.getRoom(roomID)
	if err != nil {
		return err
	}

	room.Available = available
	room.LastUpdated = time.Now()

	roomJSON, err := json.Marshal(room)
	if err != nil {
		return err
	}

	pipe := s.redis.Pipeline()
	pipe.HSet(s.ctx, "rooms", room.ID, roomJSON)
	if available {
		pipe.SAdd(s.ctx, "available_rooms", room.ID)
	} else {
		pipe.SRem(s.ctx, "available_rooms", room.ID)
	}
	_, err = pipe.Exec(s.ctx)

	if err == nil {
		s.publishRoomUpdate(*room)
	}

	return err
}

// Booking System
func (s *Server) handleBooking(w http.ResponseWriter, r *http.Request) {
	var bookingReq BookingRequest
	if err := json.NewDecoder(r.Body).Decode(&bookingReq); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create distributed lock for the room
	lock := NewDistributedLock(s.redis, bookingReq.RoomID, 30*time.Second)
	
	// Try to acquire lock
	acquired, err := lock.Acquire()
	if err != nil {
		http.Error(w, "Error acquiring lock", http.StatusInternalServerError)
		return
	}
	if !acquired {
		http.Error(w, "Room is currently being booked by another user", http.StatusConflict)
		return
	}
	defer lock.Release()

	// Check if room is available
	room, err := s.getRoom(bookingReq.RoomID)
	if err != nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}
	if !room.Available {
		http.Error(w, "Room is not available", http.StatusBadRequest)
		return
	}

	// Parse dates
	checkIn, err := time.Parse(time.RFC3339, bookingReq.CheckIn)
	if err != nil {
		http.Error(w, "Invalid check-in date", http.StatusBadRequest)
		return
	}
	checkOut, err := time.Parse(time.RFC3339, bookingReq.CheckOut)
	if err != nil {
		http.Error(w, "Invalid check-out date", http.StatusBadRequest)
		return
	}

	// Create booking
	booking := Booking{
		ID:        fmt.Sprintf("booking:%d", time.Now().UnixNano()),
		RoomID:    bookingReq.RoomID,
		UserID:    bookingReq.UserID,
		CheckIn:   checkIn,
		CheckOut:  checkOut,
		Status:    "confirmed",
		CreatedAt: time.Now(),
	}

	// Save booking and update room availability
	pipe := s.redis.Pipeline()
	bookingJSON, _ := json.Marshal(booking)
	pipe.Set(s.ctx, booking.ID, bookingJSON, 0)
	pipe.SAdd(s.ctx, fmt.Sprintf("user_bookings:%s", booking.UserID), booking.ID)
	_, err = pipe.Exec(s.ctx)

	if err != nil {
		http.Error(w, "Error creating booking", http.StatusInternalServerError)
		return
	}

	// Update room availability
	err = s.updateRoomAvailability(booking.RoomID, false)
	if err != nil {
		http.Error(w, "Error updating room availability", http.StatusInternalServerError)
		return
	}

	// Publish booking event
	s.publishBookingUpdate(booking)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(booking)
}

// WebSocket Management
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	s.clients[conn] = true
	defer delete(s.clients, conn)

	// Subscribe to Redis pub/sub channels
	pubsub := s.redis.Subscribe(s.ctx, "room_updates", "booking_updates")
	defer pubsub.Close()

	// Handle incoming messages
	go func() {
		for {
			msg, err := pubsub.ReceiveMessage(s.ctx)
			if err != nil {
				log.Printf("Error receiving message: %v", err)
				return
			}

			err = conn.WriteMessage(websocket.TextMessage, []byte(msg.Payload))
			if err != nil {
				log.Printf("Error writing message: %v", err)
				return
			}
		}
	}()

	// Keep connection alive
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

// Pub/Sub Methods
func (s *Server) publishRoomUpdate(room Room) {
	update := map[string]interface{}{
		"type": "room_update",
		"data": room,
	}
	updateJSON, _ := json.Marshal(update)
	s.redis.Publish(s.ctx, "room_updates", updateJSON)
}

func (s *Server) publishBookingUpdate(booking Booking) {
	update := map[string]interface{}{
		"type": "booking_update",
		"data": booking,
	}
	updateJSON, _ := json.Marshal(update)
	s.redis.Publish(s.ctx, "booking_updates", updateJSON)
}

func main() {
	server := NewServer()
	router := mux.NewRouter()

	router.HandleFunc("/api/rooms", server.handleCreateRoom).Methods("POST")
	router.HandleFunc("/api/rooms/{id}", server.handleGetRoom).Methods("GET")
	router.HandleFunc("/api/bookings", server.handleBooking).Methods("POST")
	router.HandleFunc("/ws", server.handleWebSocket)

	log.Fatal(http.ListenAndServe(":8080", router))
}
