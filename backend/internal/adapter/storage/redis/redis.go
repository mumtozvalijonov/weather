package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
}

func New(ctx context.Context) (*Redis, error) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
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

func (r *Redis) Get(ctx context.Context, key string) ([]byte, error) {
	res, err := r.client.Get(ctx, key).Result()
	bytes := []byte(res)
	return bytes, err
}

func (r *Redis) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

func (r *Redis) AddToGeoData(ctx context.Context, geoKey, locationName string, longitude, latitude float64) error {
	return r.client.GeoAdd(ctx, geoKey, &redis.GeoLocation{
		Name:      locationName,
		Longitude: longitude,
		Latitude:  latitude,
	}).Err()
}

func (r *Redis) FindKeyWithinRadius(ctx context.Context, geoKey string, longitude, latitude, radius float64) (string, error) {
	results, err := r.client.GeoSearch(ctx, geoKey, &redis.GeoSearchQuery{
		Longitude:  longitude,
		Latitude:   latitude,
		Radius:     radius,
		RadiusUnit: "m",
		Count:      1,
	}).Result()
	if err != nil {
		return "", nil
	}

	if len(results) > 0 {
		return results[0], nil
	}
	return "", nil
}
