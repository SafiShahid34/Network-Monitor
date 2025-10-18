package storage

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// RedisClient wraps the Redis client
type RedisClient struct {
	Client *redis.Client
}

// New initializes and returns a Redis client
func New(addr string) *RedisClient {
	return &RedisClient{
		Client: redis.NewClient(&redis.Options{
			Addr: addr,
		}),
	}
}

// Close shuts down the client
func (r *RedisClient) Close() error {
	return r.Client.Close()
}

// Ping tests the Redis connection
func (r *RedisClient) Ping(ctx context.Context) error {
	return r.Client.Ping(ctx).Err()
}
