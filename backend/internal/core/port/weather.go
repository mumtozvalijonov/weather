package port

import (
	"context"

	"github.com/mumtozvalijonov/weather/internal/core/domain"
)

type WeatherProvider interface {
	GetForecast(ctx context.Context, req domain.ForecastRequest) (domain.Forecast, error)
}

type WeatherService interface {
	GetForecast(ctx context.Context, req domain.ForecastRequest) (domain.Forecast, error)
}
