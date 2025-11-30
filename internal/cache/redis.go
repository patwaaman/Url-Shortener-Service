package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
}

func NewRedis(addr, password string, db int) *RedisClient {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &RedisClient{client: rdb}
}

func (r *RedisClient) GetURL(ctx context.Context, code string) (string, error) {
	key := "short:" + code
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

func (r *RedisClient) SetURL(ctx context.Context, code, original string, ttl time.Duration) error {
	key := "short:" + code
	return r.client.Set(ctx, key, original, ttl).Err()
}

func (r *RedisClient) Close() error {
	return r.client.Close()
}
