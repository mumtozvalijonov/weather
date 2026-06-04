package service

import (
	"context"
	"fmt"
	"time"

	"github.com/mmcloughlin/geohash"
	"github.com/mumtozvalijonov/weather/internal/core/domain"
	"github.com/mumtozvalijonov/weather/internal/core/port"
	"golang.org/x/sync/singleflight"
)

type WeatherService struct {
	requestGroup singleflight.Group
	provider     port.WeatherProvider
	weatherRepo  port.WeatherRepository
}

func NewWeatherService(weatherProvider port.WeatherProvider, weatherRepository port.WeatherRepository) *WeatherService {
	return &WeatherService{
		provider:    weatherProvider,
		weatherRepo: weatherRepository,
	}
}

func (s *WeatherService) fetchForecastFromProvider(ctx context.Context, req domain.ForecastRequest) (*domain.Forecast, error) {
	result, err := s.provider.GetForecast(ctx, req)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *WeatherService) GetForecast(ctx context.Context, req domain.ForecastRequest) (*domain.Forecast, error) {
	areaKey, err := s.makeAreaKey(ctx, req)
	if err != nil {
		return nil, ErrForecastUnavailable
	}

	result, err, _ := s.requestGroup.Do(areaKey, func() (any, error) {
		cachedData, err := s.weatherRepo.Get(ctx, areaKey)
		if err == nil {
			return cachedData, nil
		}

		// If we end up here, means data not found in cache.
		// NOTE 1: Because we don't retire members of GeoSet
		// we might be requesting forecast for a location within 5km
		// from the position the areaKey belongs to
		providerResult, err := s.fetchForecastFromProvider(ctx, req)
		if err != nil {
			return nil, err
		}

		// Note 2: Because of `Note 1` -> we will now store data
		// related to a potentially different location, but not much distant
		_ = s.weatherRepo.Set(ctx, areaKey, providerResult, time.Hour)
		return providerResult, nil
	})

	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrForecastUnavailable, err)
	}

	if forecast, ok := result.(*domain.Forecast); ok {
		return forecast, nil
	}
	return nil, ErrForecastUnavailable
}

func (s *WeatherService) makeAreaKey(ctx context.Context, req domain.ForecastRequest) (string, error) {
	locationName := makeCacheKey(req)
	return s.weatherRepo.FindKeyWithinRadiusWithUpsert(ctx, "forecast", locationName, req.Longitude, req.Latitude, 5000)
}

func makeCacheKey(req domain.ForecastRequest) string {
	return geohash.Encode(req.Latitude, req.Longitude)
}
