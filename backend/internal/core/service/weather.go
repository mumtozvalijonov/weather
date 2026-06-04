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

	cachedData, err := s.weatherRepo.Get(ctx, cacheKey)
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

func (s *WeatherService) fetchForecastFromProvider(ctx context.Context, req domain.ForecastRequest) (*domain.Forecast, error) {
	result, err := s.provider.GetForecast(ctx, req)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *WeatherService) tryPersistForecastInRepo(ctx context.Context, forecast *domain.Forecast) bool {
	cacheKey := makeCacheKey(domain.ForecastRequest{
		Latitude:  forecast.Latitude,
		Longitude: forecast.Longitude,
	})
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return s.weatherRepo.Set(ctx, cacheKey, forecast, time.Hour)
	})
	g.Go(func() error {
		return s.weatherRepo.AddGeoData(ctx, "forecast", cacheKey, forecast.Longitude, forecast.Latitude)
	})
	err := g.Wait()
	return err == nil
}

func (s *WeatherService) GetForecast(ctx context.Context, req domain.ForecastRequest) (*domain.Forecast, error) {
	forecast, ok := s.tryGetForecastFromRepo(ctx, req)
	if ok {
		return forecast, nil
	}

	// TODO: singleflight here

	result, err := s.fetchForecastFromProvider(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrForecastUnavailable, err)
	}

	s.tryPersistForecastInRepo(ctx, result)
	return result, nil
}

func (s *WeatherService) getCacheKey(ctx context.Context, req domain.ForecastRequest) (string, error) {
	return s.weatherRepo.FindKeyWithinRadius(ctx, "forecast", req.Longitude, req.Latitude, 5000)
}

func makeCacheKey(req domain.ForecastRequest) string {
	return geohash.Encode(req.Latitude, req.Longitude)
}
