package domain

import "time"

type Reservation struct {
    TicketID    string    `json:"ticket_id"`
    UserID      string    `json:"user_id"`
    CreatedAt   time.Time `json:"created_at"`
    ExpiresAt   time.Time `json:"expires_at"`
}

type ReservationRequest struct {
    TicketID    string `json:"ticket_id"`
    UserID      string `json:"user_id"`
}