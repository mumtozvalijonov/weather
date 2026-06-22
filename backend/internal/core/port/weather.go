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

type WeatherRepository interface {
	Get(ctx context.Context, key string) (*domain.Forecast, error)
	Set(ctx context.Context, key string, value *domain.Forecast, ttl time.Duration) error
}
