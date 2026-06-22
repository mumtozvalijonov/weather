package service_test

import (
	"context"
	"sync"
	"time"

	"github.com/mumtozvalijonov/weather/internal/core/domain"
	"github.com/mumtozvalijonov/weather/internal/core/service"
)

type fakeWeatherRepository struct {
	mu   sync.Mutex
	data map[string]*domain.Forecast
}

func newFakeWeatherRepository() *fakeWeatherRepository {
	return &fakeWeatherRepository{
		data: make(map[string]*domain.Forecast),
	}
}

func (f *fakeWeatherRepository) Get(ctx context.Context, key string) (*domain.Forecast, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	value, ok := f.data[key]
	if !ok {
		return nil, service.ErrForecastUnavailable
	}
	return value, nil
}

func (f *fakeWeatherRepository) Set(ctx context.Context, key string, value *domain.Forecast, ttl time.Duration) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.data[key] = value
	return nil
}
