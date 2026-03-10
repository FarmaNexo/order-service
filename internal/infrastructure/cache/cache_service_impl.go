package cache

import (
	"context"
	"time"

	"github.com/farmanexo/order-service/internal/domain/services"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type RedisCacheService struct {
	redisClient *RedisClient
	logger      *zap.Logger
}

func NewRedisCacheService(redisClient *RedisClient, logger *zap.Logger) *RedisCacheService {
	return &RedisCacheService{redisClient: redisClient, logger: logger}
}

func (c *RedisCacheService) Get(ctx context.Context, key string) (string, error) {
	val, err := c.redisClient.Client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

func (c *RedisCacheService) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return c.redisClient.Client.Set(ctx, key, value, ttl).Err()
}

func (c *RedisCacheService) Delete(ctx context.Context, key string) error {
	return c.redisClient.Client.Del(ctx, key).Err()
}

func (c *RedisCacheService) DeleteByPattern(ctx context.Context, pattern string) error {
	iter := c.redisClient.Client.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		c.redisClient.Client.Del(ctx, iter.Val())
	}
	return iter.Err()
}

var _ services.CacheService = (*RedisCacheService)(nil)
