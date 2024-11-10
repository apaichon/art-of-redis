package models

import "time"

type Draw struct {
	ID            string    `json:"id"`
	Number        string    `json:"number"`
	Status        string    `json:"status"`
	WinningNumber string    `json:"winningNumber,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
	CompletedAt   time.Time `json:"completedAt,omitempty"`
}

type Winner struct {
	DrawID    string    `json:"drawId"`
	Number    string    `json:"number"`
	Prize     string    `json:"prize"`
	ClaimedAt time.Time `json:"claimedAt"`
}

type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}