package cache

import (
	"context"
	"fmt"

	"github.com/farmanexo/order-service/pkg/config"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type RedisClient struct {
	Client *redis.Client
	logger *zap.Logger
}

func NewRedisClient(cfg config.RedisConfig, logger *zap.Logger) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:       fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:   cfg.Password,
		DB:         cfg.DB,
		MaxRetries: cfg.MaxRetries,
		PoolSize:   cfg.PoolSize,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		logger.Warn("Redis no disponible, continuando sin cache", zap.Error(err))
	} else {
		logger.Info("Conexión a Redis establecida")
	}

	return &RedisClient{Client: client, logger: logger}, nil
}

func (r *RedisClient) Close() error {
	return r.Client.Close()
}
