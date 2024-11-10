package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"luckydraw/internal/models"
)

type RedisStore struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisStore(addr string) (*RedisStore, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   0,
	})
	
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	
	return &RedisStore{
		client: client,
		ctx:    ctx,
	}, nil
}

func (s *RedisStore) StoreDraw(draw *models.Draw) error {
	drawJSON, err := json.Marshal(draw)
	if err != nil {
		return err
	}
	
	return s.client.Set(s.ctx, draw.ID, drawJSON, 24*time.Hour).Err()
}

func (s *RedisStore) GetDraw(id string) (*models.Draw, error) {
	drawJSON, err := s.client.Get(s.ctx, id).Result()
	if err != nil {
		return nil, err
	}
	
	var draw models.Draw
	if err := json.Unmarshal([]byte(drawJSON), &draw); err != nil {
		return nil, err
	}
	
	return &draw, nil
}

func (s *RedisStore) StoreWinner(winner *models.Winner) error {
	winnerJSON, err := json.Marshal(winner)
	if err != nil {
		return err
	}
	
	pipe := s.client.Pipeline()
	pipe.Set(s.ctx, fmt.Sprintf("winner:%s", winner.Number), winnerJSON, 30*24*time.Hour)
	pipe.SAdd(s.ctx, "claimed_prizes", winner.Number)
	_, err = pipe.Exec(s.ctx)
	
	return err
}

func (s *RedisStore) IsNumberClaimed(number string) (bool, error) {
	return s.client.SIsMember(s.ctx, "claimed_prizes", number).Result()
}

func (s *RedisStore) ClaimPrize(drawID, userID string) error {
	// Check if the draw exists
	exists, err := s.client.Exists(s.ctx, drawID).Result()
	if err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("draw ID %s does not exist", drawID)
	}

	// Update the draw record with the user ID
	err = s.client.HSet(s.ctx, drawID, userID, true).Err()
	if err != nil {
		return err
	}

	return nil // Successfully claimed the prize
} 