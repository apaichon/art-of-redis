package models

import "time"
type Player struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Score     float64   `json:"score"`
	Rank      int       `json:"rank"`
	UpdatedAt time.Time `json:"updated_at"`
}
