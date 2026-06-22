package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/mmcloughlin/geohash"
	"github.com/mumtozvalijonov/weather/internal/core/domain"
	"github.com/mumtozvalijonov/weather/internal/core/port"
	"golang.org/x/sync/singleflight"
)

type WeatherService struct {
	flightGroup *singleflight.Group
	provider    port.WeatherProvider
	weatherRepo port.WeatherRepository
}

func NewWeatherService(
	weatherProvider port.WeatherProvider,
	weatherRepo port.WeatherRepository,
) *WeatherService {
	return &WeatherService{
		provider:    weatherProvider,
		weatherRepo: weatherRepo,
		flightGroup: new(singleflight.Group),
	}
}

func (s *WeatherService) GetForecast(ctx context.Context, req domain.ForecastRequest) (*domain.Forecast, error) {
	cacheKey := makeCacheKey(req)

	result, err, _ := s.flightGroup.Do(cacheKey, func() (any, error) {
		if cachedData, err := s.weatherRepo.Get(ctx, cacheKey); err == nil {
			return cachedData, nil
		}

		result, err := s.fetchForecastFromProvider(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrForecastUnavailable, err)
		}

		s.weatherRepo.Set(ctx, cacheKey, result, time.Hour)
		return result, nil
	})

	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrForecastUnavailable, err)
	}

	forecast, ok := result.(*domain.Forecast)
	if !ok {
		return nil, ErrForecastUnavailable
	}

	return forecast, nil
}

func (s *WeatherService) fetchForecastFromProvider(ctx context.Context, req domain.ForecastRequest) (*domain.Forecast, error) {
	result, err := s.provider.GetForecast(ctx, req)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func makeCacheKey(req domain.ForecastRequest) string {
	return geohash.Encode(
		math.Round(req.Latitude*100)/100,
		math.Round(req.Longitude*100)/100,
	)
}
