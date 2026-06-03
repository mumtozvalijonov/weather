package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/mmcloughlin/geohash"
	"github.com/mumtozvalijonov/weather/internal/core/domain"
	"github.com/mumtozvalijonov/weather/internal/core/port"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/singleflight"
)

type WeatherService struct {
	requestGroup singleflight.Group
	provider     port.WeatherProvider
	cache        port.CacheRepository
}

func NewWeatherService(weatherProvider port.WeatherProvider, cache port.CacheRepository) *WeatherService {
	return &WeatherService{
		provider: weatherProvider,
		cache:    cache,
	}
}

func (s *WeatherService) GetForecast(ctx context.Context, req domain.ForecastRequest) (domain.Forecast, error) {
	cacheKey, err := s.getCacheKey(ctx, req)
	if err != nil {
		return domain.Forecast{}, err
	}

	if cacheKey != "" {
		cachedData, err := s.cache.Get(ctx, cacheKey)
		if err == nil {
			var result domain.Forecast
			err := json.Unmarshal(cachedData, &result)
			if err != nil {
				return domain.Forecast{}, err
			}
			return result, nil
		}
	}
	cacheKey = makeCacheKey(req)

	// TODO: singleflight here

	result, err := s.provider.GetForecast(ctx, req)
	if err != nil {
		return domain.Forecast{}, err
	}

	resultSerialized, err := json.Marshal(result)
	if err != nil {
		return domain.Forecast{}, err
	}

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return s.cache.Set(ctx, cacheKey, resultSerialized, time.Hour)
	})

	g.Go(func() error {
		return s.cache.AddToGeoData(ctx, "forecast", cacheKey, req.Longitude, req.Latitude)
	})

	if err := g.Wait(); err != nil {
		return domain.Forecast{}, err
	}

	return result, nil
}

func (s *WeatherService) getCacheKey(ctx context.Context, req domain.ForecastRequest) (string, error) {
	return s.cache.FindKeyWithinRadius(ctx, "forecast", req.Longitude, req.Latitude, 5000)
}

func makeCacheKey(req domain.ForecastRequest) string {
	return geohash.Encode(req.Latitude, req.Longitude)
}
