package redis

import (
    "context"
    "ticket-reservation/config"

    "github.com/redis/go-redis/v9"
)

type Client struct {
    *redis.Client
}

func NewClient(cfg *config.Config) (*Client, error) {
    client := redis.NewClient(&redis.Options{
        Addr:     cfg.RedisAddr,
        Password: cfg.RedisPassword,
        DB:       cfg.RedisDB,
    })

    // Test connection
    ctx := context.Background()
    if err := client.Ping(ctx).Err(); err != nil {
        return nil, err
    }

    // Enable keyspace notifications for expiration events
    client.ConfigSet(ctx, "notify-keyspace-events", "Ex")

    return &Client{client}, nil
}

func (c *Client) Close() error {
    return c.Client.Close()
}