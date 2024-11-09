package repository

import (
    "context"
    "encoding/json"
    "fmt"

    "github.com/go-redis/redis/v8"
    "leaderboard/internal/config"
    "leaderboard/internal/domain/models"
)

type RedisRepository struct {
    client *redis.Client
    config *config.Config
}

func NewRedisRepository(cfg *config.Config) (*RedisRepository, error) {
    client := redis.NewClient(&redis.Options{
        Addr: cfg.RedisAddr,
        DB:   0,
    })
    
    // Test connection
    if err := client.Ping(context.Background()).Err(); err != nil {
        return nil, fmt.Errorf("redis connection failed: %w", err)
    }
    
    return &RedisRepository{
        client: client,
        config: cfg,
    }, nil
}

func (r *RedisRepository) UpdateScore(ctx context.Context, player *models.Player) error {
    pipe := r.client.Pipeline()
    
    // Update score in sorted set
    pipe.ZAdd(ctx, r.config.LeaderboardKey, &redis.Z{
        Score:  player.Score,
        Member: player.ID,
    })
    
    // Store player details
    playerData, err := json.Marshal(player)
    if err != nil {
        return fmt.Errorf("failed to marshal player: %w", err)
    }
    
    pipe.Set(ctx, fmt.Sprintf("player:%s", player.ID), playerData, 0)
    
    _, err = pipe.Exec(ctx)
    return err
}

func (r *RedisRepository) GetLeaderboard(ctx context.Context) ([]*models.Player, error) {
    results, err := r.client.ZRevRangeWithScores(ctx, r.config.LeaderboardKey, 0, 99).Result()
    if err != nil {
        return nil, fmt.Errorf("failed to get leaderboard: %w", err)
    }

    var players []*models.Player
    for rank, z := range results {
        playerID := z.Member.(string)
        playerData, err := r.client.Get(ctx, fmt.Sprintf("player:%s", playerID)).Result()
        if err != nil {
            continue
        }

        var player models.Player
        if err := json.Unmarshal([]byte(playerData), &player); err != nil {
            continue
        }
        
        player.Rank = rank + 1
        player.Score = z.Score
        players = append(players, &player)
    }

    return players, nil
}

func (r *RedisRepository) ClearData(ctx context.Context) error {
    // Clear all data in the current Redis database
    if err := r.client.FlushDB(ctx).Err(); err != nil {
        return fmt.Errorf("failed to clear Redis data: %w", err)
    }
    return nil
}