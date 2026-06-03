package port

import (
	"context"
	"time"
)

type CacheRepository interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	AddToGeoData(ctx context.Context, geoKey, locationName string, longitude, latitude float64) error
	FindKeyWithinRadius(ctx context.Context, geoKey string, longitude, latitude, radius float64) (string, error)
	Close() error
}
