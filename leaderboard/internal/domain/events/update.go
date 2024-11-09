package events

import (
    "time"
    "leaderboard/internal/domain/models"
)

type LeaderboardUpdate struct {
    Type      string           `json:"type"`
    Player    *models.Player   `json:"player,omitempty"`
    Rankings  []*models.Player `json:"rankings,omitempty"`
    Timestamp int64           `json:"timestamp"`
}

func NewUpdate(updateType string, player *models.Player, rankings []*models.Player) LeaderboardUpdate {
    return LeaderboardUpdate{
        Type:      updateType,
        Player:    player,
        Rankings:  rankings,
        Timestamp: time.Now().Unix(),
    }
}