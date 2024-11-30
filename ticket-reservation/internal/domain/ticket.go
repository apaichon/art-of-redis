package domain

import "time"

type Ticket struct {
    ID          string    `json:"id"`
    Reserved    bool      `json:"reserved"`
    ReservedBy  string    `json:"reserved_by,omitempty"`
    ReservedAt  time.Time `json:"reserved_at,omitempty"`
    ExpiresAt   time.Time `json:"expires_at,omitempty"`
}
