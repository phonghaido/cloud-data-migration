package handlers

import (
	"context"

	"github.com/phonghaido/cloud-data-migration/internal/config"
	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	RedisConfig config.RedisConfig
	RedisClient *redis.Client
}

func NewRedisClient(c config.RedisConfig) RedisClient {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     c.Address,
		Password: c.Password,
		DB:       0,
	})

	return RedisClient{
		RedisConfig: c,
		RedisClient: redisClient,
	}
}

func (r RedisClient) IsPublished(ctx context.Context, key string) (int64, error) {
	return r.RedisClient.Exists(ctx, "published:"+key).Result()
}

func (r RedisClient) IsConsumed(ctx context.Context, key string) (int64, error) {
	return r.RedisClient.Exists(ctx, "consumed:"+key).Result()
}

func (r RedisClient) MarkAsPublished(ctx context.Context, key string, value string) error {
	return r.RedisClient.Set(ctx, "published:"+key, value, 0).Err()
}

func (r RedisClient) MarkAsConsumed(ctx context.Context, key string, value string) error {
	return r.RedisClient.Set(ctx, "consumed:"+key, value, 0).Err()
}
