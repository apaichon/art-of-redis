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

type TicketSale struct {
	ID          string    `json:"id"`
	ConcertID   string    `json:"concertId"`
	ConcertName string    `json:"concertName"`
	Price       float64   `json:"price"`
	Category    string    `json:"category"`
	Timestamp   time.Time `json:"timestamp"`
}

type SalesAnalytics struct {
	TotalRevenue     float64                 `json:"totalRevenue"`
	TicketsSold      int                     `json:"ticketsSold"`
	SalesByCategory  map[string]int          `json:"salesByCategory"`
	RevenueByHour    []HourlyRevenue        `json:"revenueByHour"`
	TopConcerts      []ConcertSales         `json:"topConcerts"`
	RecentSales      []TicketSale           `json:"recentSales"`
	CategoryRevenue  map[string]float64      `json:"categoryRevenue"`
	HourlySalesCount map[string]int          `json:"hourlySalesCount"`
}

type HourlyRevenue struct {
	Hour     string  `json:"hour"`
	Revenue  float64 `json:"revenue"`
	SaleCount int    `json:"saleCount"`
}

type ConcertSales struct {
	ConcertID   string  `json:"concertId"`
	ConcertName string  `json:"concertName"`
	Revenue     float64 `json:"revenue"`
	TicketsSold int     `json:"ticketsSold"`
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

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	s.clients[conn] = true
	defer delete(s.clients, conn)

	// Send initial analytics
	analytics, err := s.getAnalytics()
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

func (s *Server) recordTicketSale(sale TicketSale) error {
	// Store sale in Redis
	saleJSON, _ := json.Marshal(sale)
	pipe := s.redis.Pipeline()

	// Store complete sale object in a list
	pipe.LPush(s.ctx, "recent_sales", saleJSON)
	pipe.LTrim(s.ctx, "recent_sales", 0, 99) // Keep last 100 sales

	// Increment total sales count
	pipe.Incr(s.ctx, "total_tickets_sold")

	// Add to total revenue
	pipe.IncrByFloat(s.ctx, "total_revenue", sale.Price)

	// Increment category count
	pipe.Incr(s.ctx, fmt.Sprintf("category:%s:count", sale.Category))
	pipe.IncrByFloat(s.ctx, fmt.Sprintf("category:%s:revenue", sale.Category), sale.Price)

	// Track hourly sales
	hour := sale.Timestamp.Format("2006-01-02:15")
	pipe.Incr(s.ctx, fmt.Sprintf("hourly:%s:count", hour))
	pipe.IncrByFloat(s.ctx, fmt.Sprintf("hourly:%s:revenue", hour), sale.Price)

	// Track concert-specific stats
	pipe.Incr(s.ctx, fmt.Sprintf("concert:%s:count", sale.ConcertID))
	pipe.IncrByFloat(s.ctx, fmt.Sprintf("concert:%s:revenue", sale.ConcertID), sale.Price)
	pipe.HSet(s.ctx, "concert_names", sale.ConcertID, sale.ConcertName)

	_, err := pipe.Exec(s.ctx)
	if err != nil {
		return err
	}

	// Broadcast update to all connected clients
	analytics, err := s.getAnalytics()
	if err == nil {
		for client := range s.clients {
			client.WriteJSON(analytics)
		}
	}

	return nil
}

func (s *Server) getAnalytics() (*SalesAnalytics, error) {
	analytics := &SalesAnalytics{
		SalesByCategory:  make(map[string]int),
		CategoryRevenue:  make(map[string]float64),
		HourlySalesCount: make(map[string]int),
	}

	// Get total revenue and tickets sold
	totalRevenue, err := s.redis.Get(s.ctx, "total_revenue").Float64()
	if err != nil && err != redis.Nil {
		return nil, err
	}
	analytics.TotalRevenue = totalRevenue

	ticketsSold, err := s.redis.Get(s.ctx, "total_tickets_sold").Int()
	if err != nil && err != redis.Nil {
		return nil, err
	}
	analytics.TicketsSold = ticketsSold

	// Get recent sales
	recentSalesJSON, err := s.redis.LRange(s.ctx, "recent_sales", 0, 9).Result()
	if err != nil {
		return nil, err
	}
	for _, saleJSON := range recentSalesJSON {
		var sale TicketSale
		json.Unmarshal([]byte(saleJSON), &sale)
		analytics.RecentSales = append(analytics.RecentSales, sale)
	}

	// Get category stats
	categories := []string{"VIP", "Standard", "Economy"}
	for _, category := range categories {
		count, _ := s.redis.Get(s.ctx, fmt.Sprintf("category:%s:count", category)).Int()
		revenue, _ := s.redis.Get(s.ctx, fmt.Sprintf("category:%s:revenue", category)).Float64()
		analytics.SalesByCategory[category] = count
		analytics.CategoryRevenue[category] = revenue
	}

	// Get hourly revenue data
	now := time.Now()
	for i := 23; i >= 0; i-- {
		hour := now.Add(time.Duration(-i) * time.Hour).Format("2006-01-02:15")
		revenue, _ := s.redis.Get(s.ctx, fmt.Sprintf("hourly:%s:revenue", hour)).Float64()
		count, _ := s.redis.Get(s.ctx, fmt.Sprintf("hourly:%s:count", hour)).Int()
		analytics.RevenueByHour = append(analytics.RevenueByHour, HourlyRevenue{
			Hour:      hour[11:], // Extract just the hour
			Revenue:   revenue,
			SaleCount: count,
		})
	}

	// Get top concerts
	concertIDs, _ := s.redis.HGetAll(s.ctx, "concert_names").Result()
	var concerts []ConcertSales

	for id, name := range concertIDs {
		revenue, _ := s.redis.Get(s.ctx, fmt.Sprintf("concert:%s:revenue", id)).Float64()
		count, _ := s.redis.Get(s.ctx, fmt.Sprintf("concert:%s:count", id)).Int()
		concerts = append(concerts, ConcertSales{
			ConcertID:   id,
			ConcertName: name,
			Revenue:     revenue,
			TicketsSold: count,
		})
	}

	// Sort concerts by revenue
	analytics.TopConcerts = concerts

	return analytics, nil
}

func (s *Server) handleSale(w http.ResponseWriter, r *http.Request) {
	var sale TicketSale
	if err := json.NewDecoder(r.Body).Decode(&sale); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sale.Timestamp = time.Now()
	if err := s.recordTicketSale(sale); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func main() {
	server := NewServer()
	router := mux.NewRouter()

	router.HandleFunc("/ws", server.handleWebSocket)
	router.HandleFunc("/api/sales", server.handleSale).Methods("POST")

	log.Fatal(http.ListenAndServe(":8080", router))
}
