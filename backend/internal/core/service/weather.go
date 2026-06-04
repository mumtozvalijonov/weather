package service

import (
	"context"
	"fmt"
	"time"

	"github.com/mmcloughlin/geohash"
	"github.com/mumtozvalijonov/weather/internal/core/domain"
	"github.com/mumtozvalijonov/weather/internal/core/port"
	"golang.org/x/sync/errgroup"
)

type WeatherService struct {
	provider    port.WeatherProvider
	weatherRepo port.WeatherRepository
}

func NewWeatherService(weatherProvider port.WeatherProvider, weatherRepository port.WeatherRepository) *WeatherService {
	return &WeatherService{
		provider:    weatherProvider,
		weatherRepo: weatherRepository,
	}
}

func (s *WeatherService) tryGetForecastFromRepo(ctx context.Context, req domain.ForecastRequest) (*domain.Forecast, bool) {
	cacheKey, err := s.getCacheKey(ctx, req)
	if err != nil {
		return nil, false
	}

	if cacheKey == "" {
		return nil, false
	}

	cachedData, err := s.weatherRepo.Get(ctx, cacheKey);
	if err == nil {
		return cachedData, true
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		_ = s.weatherRepo.DeleteGeoData(ctx, "forecast", cacheKey)
	}()
	return nil, false
}

func (s *WeatherService) GetForecast(ctx context.Context, req domain.ForecastRequest) (*domain.Forecast, error) {
	if forecast, ok  := s.tryGetForecastFromRepo(ctx, req); ok  {
		return forecast, nil
	}

	cacheKey := makeCacheKey(req)
	// TODO: singleflight here

	result, err := s.provider.GetForecast(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrForecastUnavailable, err)
	}

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return s.weatherRepo.Set(ctx, cacheKey, result, time.Hour)
	})
	g.Go(func() error {
		return s.weatherRepo.AddGeoData(ctx, "forecast", cacheKey, req.Longitude, req.Latitude)
	})
	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrForecastUnavailable, err)
	}

	return result, nil
}

func (s *WeatherService) getCacheKey(ctx context.Context, req domain.ForecastRequest) (string, error) {
	return s.weatherRepo.FindKeyWithinRadius(ctx, "forecast", req.Longitude, req.Latitude, 5000)
}

func makeCacheKey(req domain.ForecastRequest) string {
	return geohash.Encode(req.Latitude, req.Longitude)
}
