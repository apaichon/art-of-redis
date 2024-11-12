package ports

import (
	"context"
	"leaderboard/internal/domain/models"
)

type LeaderboardRepository interface {
	UpdateScore(ctx context.Context, player *models.Player) error
	GetLeaderboard(ctx context.Context) ([]*models.Player, error)
	RemoveLeaderboard(ctx context.Context, key string) error
}

type LeaderboardService interface {
	UpdatePlayerScore(ctx context.Context, player *models.Player) error
	GetRankings(ctx context.Context) ([]*models.Player, error)
}
