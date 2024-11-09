package service

import (
    "context"
    
    "leaderboard/internal/domain/models"
    "leaderboard/internal/ports"
)

type LeaderboardService struct {
    repo ports.LeaderboardRepository
}

func NewLeaderboardService(repo ports.LeaderboardRepository) *LeaderboardService {
    return &LeaderboardService{
        repo: repo,
    }
}

func (s *LeaderboardService) UpdatePlayerScore(ctx context.Context, player *models.Player) error {
    return s.repo.UpdateScore(ctx, player)
}

func (s *LeaderboardService) GetRankings(ctx context.Context) ([]*models.Player, error) {
    return s.repo.GetLeaderboard(ctx)
}