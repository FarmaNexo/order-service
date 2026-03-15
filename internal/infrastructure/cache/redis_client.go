package cache

import (
	"context"
	"crypto/tls"

	"github.com/farmanexo/order-service/pkg/config"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type RedisClient struct {
	Client *redis.Client
	logger *zap.Logger
}

func NewRedisClient(cfg config.RedisConfig, environment string, logger *zap.Logger) (*RedisClient, error) {
	opts := &redis.Options{
		Addr:       cfg.GetAddr(),
		Password:   cfg.Password,
		DB:         cfg.DB,
		MaxRetries: cfg.MaxRetries,
		PoolSize:   cfg.PoolSize,
	}

	// TLS requerido en ambientes AWS (development, production)
	if environment != "local" {
		opts.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
		logger.Info("Redis TLS habilitado", zap.String("environment", environment))
	}

	client := redis.NewClient(opts)

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
