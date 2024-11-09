package models

import "time"

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
	RevenueByHour    []HourlyRevenue         `json:"revenueByHour"`
	TopConcerts      []ConcertSales          `json:"topConcerts"`
	RecentSales      []TicketSale            `json:"recentSales"`
	CategoryRevenue  map[string]float64      `json:"categoryRevenue"`
	HourlySalesCount map[string]int          `json:"hourlySalesCount"`
}

type HourlyRevenue struct {
	Hour      string  `json:"hour"`
	Revenue   float64 `json:"revenue"`
	SaleCount int     `json:"saleCount"`
}

type ConcertSales struct {
	ConcertID   string  `json:"concertId"`
	ConcertName string  `json:"concertName"`
	Revenue     float64 `json:"revenue"`
	TicketsSold int     `json:"ticketsSold"`
}