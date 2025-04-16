package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/phonghaido/cloud-data-migration/pkg/http_error"
	"github.com/redis/go-redis/v9"
)

type SubHTTPHandler struct {
	AWSClient    AWSClient
	PubSubClient PubSubClient
	RedisClient  RedisClient
}

func NewSubHTTPHandler(aws AWSClient, pubsub PubSubClient, redis RedisClient) SubHTTPHandler {
	return SubHTTPHandler{
		AWSClient:    aws,
		PubSubClient: pubsub,
		RedisClient:  redis,
	}
}

func findRedisKeys(ctx context.Context, rds *redis.Client, prefix string) ([]string, error) {
	var (
		cursor uint64
		keys   []string
		err    error
	)

	for {
		var tmpKeys []string
		tmpKeys, cursor, err = rds.Scan(ctx, cursor, prefix, 0).Result()
		if err != nil {
			return nil, err
		}

		keys = append(keys, tmpKeys...)

		if cursor == 0 {
			break
		}
	}

	return keys, nil
}

func fromKeysToMessages(ctx context.Context, rds *redis.Client, prefix string) (map[string]Message, error) {
	keys, err := findRedisKeys(ctx, rds, prefix)
	if err != nil {
		return nil, err
	}

	result := make(map[string]Message)
	for _, key := range keys {
		val, err := rds.Get(ctx, key).Result()
		if err == redis.Nil {
			continue
		} else if err != nil {
			return nil, err
		} else {
			var msg Message
			if err := json.Unmarshal([]byte(val), &msg); err != nil {
				return nil, err
			}
			result[key] = msg
		}
	}
	return result, nil
}

func (s SubHTTPHandler) HandlerFindAll(c echo.Context) error {
	result, err := fromKeysToMessages(c.Request().Context(), s.RedisClient.RedisClient, "*")
	if err != nil {
		return err
	}
	return http_error.WriteJSON(c, http.StatusOK, result)
}

func (s SubHTTPHandler) HandlerFindRecordByType(c echo.Context) error {
	cacheType := c.QueryParam("value")
	result, err := fromKeysToMessages(c.Request().Context(), s.RedisClient.RedisClient, fmt.Sprintf("%s:*", cacheType))
	if err != nil {
		return err
	}
	return http_error.WriteJSON(c, http.StatusOK, result)
}

func (s SubHTTPHandler) HandlerFindRecordByName(c echo.Context) error {
	cacheName := c.QueryParam("value")
	result, err := fromKeysToMessages(c.Request().Context(), s.RedisClient.RedisClient, cacheName)
	if err != nil {
		return err
	}
	return http_error.WriteJSON(c, http.StatusOK, result)
}

func (s SubHTTPHandler) HandlerDeleteAll(c echo.Context) error {
	if err := s.RedisClient.RedisClient.FlushAll(c.Request().Context()).Err(); err != nil {
		return err
	}
	return http_error.WriteJSON(c, http.StatusOK, "Successfully deleted all caches")
}

func (s SubHTTPHandler) HandlerDeleteRecordByType(c echo.Context) error {
	cacheType := c.QueryParam("value")
	keys, err := findRedisKeys(c.Request().Context(), s.RedisClient.RedisClient, fmt.Sprintf("%s:*", cacheType))
	if err != nil {
		return err
	}
	if len(keys) > 0 {
		if err := s.RedisClient.RedisClient.Del(c.Request().Context(), keys...).Err(); err != nil {
			return err
		}
	}
	return http_error.WriteJSON(c, http.StatusOK, fmt.Sprintf("Successfully deleted all the caches with the prefix is %s", cacheType))
}

func (s SubHTTPHandler) HandlerDeleteRecordByName(c echo.Context) error {
	cacheName := c.QueryParam("value")
	keys, err := findRedisKeys(c.Request().Context(), s.RedisClient.RedisClient, cacheName)
	if err != nil {
		return err
	}
	if len(keys) > 0 {
		if err := s.RedisClient.RedisClient.Del(c.Request().Context(), keys...).Err(); err != nil {
			return err
		}
	}
	return http_error.WriteJSON(c, http.StatusOK, fmt.Sprintf("Successfully deleted the cache with the name is %s", cacheName))
}
