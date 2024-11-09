package api

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"sales-analytics/internal/models"
)

var concertNames = []string{
	"Rock Fest", "Jazz Night", "Pop Extravaganza", "Classical Evening", "Indie Showcase",
}

var categories = []string{
	"VIP", "General Admission", "Student", "Senior", "Family",
}

func (s *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	s.clients[conn] = true
	defer delete(s.clients, conn)

	// Send initial analytics
	analytics, err := s.analytics.GetAnalytics()
	if err == nil {
		conn.WriteJSON(analytics)
	}

	// Keep connection alive and handle incoming messages
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (s *Server) HandleSale(w http.ResponseWriter, r *http.Request) {
	var sale models.TicketSale
	if err := json.NewDecoder(r.Body).Decode(&sale); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sale.Timestamp = time.Now()
	if err := s.analytics.RecordTicketSale(sale); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Broadcast updated analytics to all connected clients
	if err := s.BroadcastAnalytics(); err != nil {
		log.Printf("Failed to broadcast analytics: %v", err)
	}

	w.WriteHeader(http.StatusCreated)
}

func (s *Server) HandleSaleRandom(w http.ResponseWriter, r *http.Request) {
	var sale models.TicketSale


		// Generate random TicketSale data
		sale = models.TicketSale{
			ID:          generateRandomID(),
			ConcertID:   generateRandomConcertID(),
			ConcertName: generateRandomConcertName(),
			Price:       generateRandomPrice(),
			Category:    generateRandomCategory(),
			Timestamp:   time.Now(),
		}

	// Record the ticket sale
	if err := s.analytics.RecordTicketSale(sale); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Broadcast updated analytics to all connected clients
	if err := s.BroadcastAnalytics(); err != nil {
		log.Printf("Failed to broadcast analytics: %v", err)
	}

	// Respond with the created sale
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(sale)
}

func (s *Server) HandleRandomSale(w http.ResponseWriter, r *http.Request) {
	// Generate random TicketSale data
	sale := models.TicketSale{
		ID:          generateRandomID(),
		ConcertID:   generateRandomConcertID(),
		ConcertName: generateRandomConcertName(),
		Price:       generateRandomPrice(),
		Category:    generateRandomCategory(),
		Timestamp:   time.Now(),
	}

	// Record the random ticket sale
	if err := s.analytics.RecordTicketSale(sale); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond with the created sale
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(sale)
}

func (s *Server) RemoveAll(w http.ResponseWriter, r *http.Request) {
	// Remove all ticket sales from Redis
	if err := s.analytics.RemoveAll(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond with a success message
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("All ticket sales removed successfully"))
}

func generateRandomID() string {
	return strconv.Itoa(rand.Intn(100000)) // Random ID between 0 and 99999
}

func generateRandomConcertID() string {
	return strconv.Itoa(rand.Intn(1000)) // Random concert ID between 0 and 999
}

func generateRandomConcertName() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano())) // Create a new random source
	return concertNames[r.Intn(len(concertNames))]      // Randomly select a concert name
}

func generateRandomPrice() float64 {
	return float64(rand.Intn(100)) + rand.Float64() // Random price between 0.00 and 100.99
}

func generateRandomCategory() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano())) // Create a new random source
	return categories[r.Intn(len(categories))] // Randomly select a category
}

