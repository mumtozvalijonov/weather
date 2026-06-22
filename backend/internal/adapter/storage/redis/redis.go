package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/mumtozvalijonov/weather/internal/core/domain"
	"github.com/redis/go-redis/v9"
)

type Config struct {
	Addr     string
	Password string
}

type Redis struct {
	client *redis.Client
}

func New(ctx context.Context, cfg Config) (*Redis, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       0,
	})

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	return &Redis{client: client}, nil
}

func (r *Redis) Close() error {
	return r.client.Close()
}

func (r *Redis) Get(ctx context.Context, key string) (*domain.Forecast, error) {
	res, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var result domain.Forecast
	if err = json.Unmarshal([]byte(res), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *Redis) Set(ctx context.Context, key string, value *domain.Forecast, ttl time.Duration) error {
	valueSerialized, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, valueSerialized, ttl).Err()
}
