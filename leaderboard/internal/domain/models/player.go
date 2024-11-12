package models


type Player struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Score     float64   `json:"score"`
	Rank      int       `json:"rank"`
}
