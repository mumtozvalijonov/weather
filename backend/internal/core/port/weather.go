package port

import (
	"context"
	"time"

	"github.com/mumtozvalijonov/weather/internal/core/domain"
)

type WeatherProvider interface {
	GetForecast(ctx context.Context, req domain.ForecastRequest) (*domain.Forecast, error)
}

type WeatherService interface {
	GetForecast(ctx context.Context, req domain.ForecastRequest) (*domain.Forecast, error)
}

type CancelFunc = func() error
type NextMessageFunc = func() (string, bool)

type WeatherRepository interface {
	Get(ctx context.Context, key string) (*domain.Forecast, error)
	Set(ctx context.Context, key string, value *domain.Forecast, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	AddGeoData(ctx context.Context, geoKey, locationName string, longitude, latitude float64) error
	DeleteGeoData(ctx context.Context, geoKey, locationName string) error
	FindKeyWithinRadiusWithUpsert(ctx context.Context, geoKey, locationName string, longitude, latitude, radius float64) (string, error)
	Close() error
	TryLock(ctx context.Context, key string, ttl time.Duration) (bool, error)
	Unlock(ctx context.Context, key string) error
	Subscribe(ctx context.Context, channelName string) (CancelFunc, NextMessageFunc)
	Publish(ctx context.Context, channelName string, message any) error
}
