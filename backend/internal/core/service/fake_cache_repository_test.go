package service_test

import (
	"context"
	"time"

	"github.com/mumtozvalijonov/weather/internal/core/domain"
	"github.com/mumtozvalijonov/weather/internal/core/service"
)

type fakeWeatherRepository struct {
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
	value, ok := f.data[key]
	if !ok {
		return nil, service.ErrForecastUnavailable
	}
	return value, nil
}

func (f *fakeWeatherRepository) Set(ctx context.Context, key string, value *domain.Forecast, ttl time.Duration) error {
	f.data[key] = value
	return nil
}

func (f *fakeWeatherRepository) Delete(ctx context.Context, key string) error {
	delete(f.data, key)
	return nil
}

func (f *fakeWeatherRepository) AddGeoData(ctx context.Context, geoKey, locationName string, longitude, latitude float64) error {
	if f.geo[geoKey] == nil {
		f.geo[geoKey] = map[string]fakeGeoLocation{}
	}

	f.geo[geoKey][locationName] = fakeGeoLocation{
		longitude: longitude,
		latitude:  latitude,
	}
	return nil
}

func (f *fakeWeatherRepository) DeleteGeoData(ctx context.Context, geoKey, locationName string) error {
	delete(f.geo[geoKey], locationName)
	return nil
}

func (f *fakeWeatherRepository) FindKeyWithinRadius(ctx context.Context, geoKey string, longitude, latitude, radius float64) (string, error) {
	for locationName, location := range f.geo[geoKey] {
		distance := distanceMeters(latitude, longitude, location.latitude, location.longitude)
		if distance <= radius {
			return locationName, nil
		}
	}

	return "", nil
}

func (f *fakeWeatherRepository) Close() error {
	clear(f.data)
	clear(f.geo)
	return nil
}
