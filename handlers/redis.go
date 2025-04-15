package handlers

import (
	"context"
	"encoding/json"

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

func (r RedisClient) IsPublished(ctx context.Context, key, eTag string) (bool, error) {
	value, err := r.RedisClient.Get(ctx, "published:"+key).Result()
	if err == redis.Nil {
		return false, nil
	} else if err != nil {
		return true, err
	} else {
		var msg Message
		if err := json.Unmarshal([]byte(value), &msg); err != nil {
			return false, err
		}
		return msg.ETag == eTag, nil
	}
}

func (r RedisClient) IsConsumed(ctx context.Context, key, eTag string) (bool, error) {
	value, err := r.RedisClient.Get(ctx, "consumed:"+key).Result()
	if err == redis.Nil {
		return false, nil
	} else if err != nil {
		return true, err
	} else {
		var msg Message
		if err := json.Unmarshal([]byte(value), &msg); err != nil {
			return false, err
		}
		return msg.ETag == eTag, nil
	}
}

func (r RedisClient) MarkAsPublished(ctx context.Context, value Message) error {
	msg, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.RedisClient.Set(ctx, "published:"+value.Key, msg, 0).Err()
}

func (r RedisClient) MarkAsConsumed(ctx context.Context, value Message) error {
	msg, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.RedisClient.Set(ctx, "consumed:"+value.Key, msg, 0).Err()
}
