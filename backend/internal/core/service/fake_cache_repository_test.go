package service_test

import (
	"context"
	"sync"
	"time"

	"github.com/mumtozvalijonov/weather/internal/core/domain"
	"github.com/mumtozvalijonov/weather/internal/core/port"
	"github.com/mumtozvalijonov/weather/internal/core/service"
)

type fakeWeatherRepository struct {
	mu   sync.Mutex
	data map[string]*domain.Forecast
	geo  map[string]map[string]fakeGeoLocation
}

type fakeGeoLocation struct {
	longitude float64
	latitude  float64
}

func newFakeWeatherRepository() *fakeWeatherRepository {
	return &fakeWeatherRepository{
		data: make(map[string]*domain.Forecast),
		geo:  make(map[string]map[string]fakeGeoLocation),
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

func (f *fakeWeatherRepository) Delete(ctx context.Context, key string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	delete(f.data, key)
	return nil
}

func (f *fakeWeatherRepository) AddGeoData(ctx context.Context, geoKey, locationName string, longitude, latitude float64) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.addGeoData(geoKey, locationName, longitude, latitude)
	return nil
}

func (f *fakeWeatherRepository) addGeoData(geoKey, locationName string, longitude, latitude float64) {
	if f.geo[geoKey] == nil {
		f.geo[geoKey] = map[string]fakeGeoLocation{}
	}

	f.geo[geoKey][locationName] = fakeGeoLocation{
		longitude: longitude,
		latitude:  latitude,
	}
}

func (f *fakeWeatherRepository) DeleteGeoData(ctx context.Context, geoKey, locationName string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	delete(f.geo[geoKey], locationName)
	return nil
}

func (f *fakeWeatherRepository) findKeyWithinRadius(geoKey string, longitude, latitude, radius float64) string {
	for locationName, location := range f.geo[geoKey] {
		distance := distanceMeters(latitude, longitude, location.latitude, location.longitude)
		if distance <= radius {
			return locationName
		}
	}

	return ""
}

func (f *fakeWeatherRepository) FindKeyWithinRadiusWithUpsert(ctx context.Context, geoKey, locationName string, longitude, latitude, radius float64) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	existingLocationName := f.findKeyWithinRadius(geoKey, longitude, latitude, radius)
	if existingLocationName != "" {
		return existingLocationName, nil
	}

	f.addGeoData(geoKey, locationName, longitude, latitude)
	return locationName, nil
}

func (f *fakeWeatherRepository) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	clear(f.data)
	clear(f.geo)
	return nil
}

func (f *fakeWeatherRepository) TryLock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	panic("")
}
func (f *fakeWeatherRepository) Unlock(ctx context.Context, key string) error {
	panic("")
}
func (f *fakeWeatherRepository) Subscribe(ctx context.Context, channelName string) (port.CancelFunc, port.NextMessageFunc) {
	panic("")
}
func (f *fakeWeatherRepository) Publish(ctx context.Context, channelName string, message any) error {
	panic("")
}
