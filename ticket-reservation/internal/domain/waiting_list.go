package domain

import "time"

type WaitingListItem struct {
    TicketID  string    `json:"ticket_id"`
    UserID    string    `json:"user_id"`
    AddedAt   time.Time `json:"added_at"`
    Position  int       `json:"position"`
}