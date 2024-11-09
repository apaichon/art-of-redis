package storage

import (
	"context"
	"github.com/go-redis/redis/v8"
)

type RedisStore struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisStore(addr string) *RedisStore {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   0,
	})

	return &RedisStore{
		client: client,
		ctx:    context.Background(),
	}
}

func (rs *RedisStore) Pipeline() redis.Pipeliner {
	return rs.client.Pipeline()
}

func (rs *RedisStore) Get(key string) *redis.StringCmd {
	return rs.client.Get(rs.ctx, key)
}

func (rs *RedisStore) LRange(key string, start, stop int64) *redis.StringSliceCmd {
	return rs.client.LRange(rs.ctx, key, start, stop)
}

func (rs *RedisStore) HGetAll(key string) *redis.StringStringMapCmd {
	return rs.client.HGetAll(rs.ctx, key)
}

func (rs *RedisStore) Close() error {

	return rs.client.Close()
}

func (rs *RedisStore) RemoveAll() *redis.StatusCmd {
	return rs.client.FlushDB(rs.ctx)
}



