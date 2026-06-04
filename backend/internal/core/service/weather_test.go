package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/mumtozvalijonov/weather/internal/core/domain"
	"github.com/mumtozvalijonov/weather/internal/core/service"
)

func TestWeatherService_GetForecast_CacheMiss(t *testing.T) {
	tests := []struct {
		name             string
		stubForecast     domain.Forecast
		stubError        error
		expectedForecast domain.Forecast
		expectedError    error
	}{
		{
			name: "happy path",
			stubForecast: domain.Forecast{
				Latitude:  40.7128,
				Longitude: -74.0060,
				Hourly: domain.Hourly{
					Time:        []string{"2026-06-04T00:00"},
					Temperature: []float64{21.5},
				},
			},
			stubError: nil,
			expectedForecast: domain.Forecast{
				Latitude:  40.7128,
				Longitude: -74.0060,
				Hourly: domain.Hourly{
					Time:        []string{"2026-06-04T00:00"},
					Temperature: []float64{21.5},
				},
			},
			expectedError: nil,
		},
		{
			name: "provider error",
			stubForecast: domain.Forecast{
				Latitude:  40.7128,
				Longitude: -74.0060,
				Hourly: domain.Hourly{
					Time:        []string{"2026-06-04T00:00"},
					Temperature: []float64{21.5},
				},
			},
			stubError:        errors.New("not found"),
			expectedForecast: domain.Forecast{},
			expectedError:    service.ErrForecastUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			stubWeatherProvider := newStubWeatherProvider(tt.stubForecast, tt.stubError)
			fakeWeatherRepo := newFakeWeatherRepository()

			service := service.NewWeatherService(stubWeatherProvider, fakeWeatherRepo)

			req := domain.ForecastRequest{Latitude: 40.7128, Longitude: -74.0060}
			result, err := service.GetForecast(ctx, req)

			if tt.expectedError != nil {
				if !errors.Is(err, tt.expectedError) {
					t.Fatalf("error mismatch: want %v, got %v", tt.expectedError, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if diff := cmp.Diff(tt.expectedForecast, *result); diff != "" {
				t.Fatalf("forecast mismatch (-want +got):\n%s", diff)
			}

			// check if persisted in repo
			key, err := fakeWeatherRepo.FindKeyWithinRadius(ctx, "forecast", req.Longitude, req.Latitude, 5000)
			if err != nil {
				t.Fatal(err)
			} else if key == "" {
				t.Fatalf("no key persisted in cache")
			}

			cachedData, err := fakeWeatherRepo.Get(ctx, key)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.expectedForecast, *cachedData); diff != "" {
				t.Fatalf("cached forecast mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestWeatherService_GetForecast_CacheHitWithinRadius(t *testing.T) {
	cachedLocation := domain.ForecastRequest{
		Latitude:  40.7128,
		Longitude: -74.0060,
	}
	req := domain.ForecastRequest{
		Latitude:  40.7130,
		Longitude: -74.0062,
	}

	cached := domain.Forecast{
		Latitude:  cachedLocation.Latitude,
		Longitude: cachedLocation.Longitude,
		Hourly: domain.Hourly{
			Time:        []string{"2026-06-04T00:00"},
			Temperature: []float64{18.2},
		},
	}

	providerForecast := domain.Forecast{
		Latitude:  0,
		Longitude: 0,
		Hourly: domain.Hourly{
			Time:        []string{"provider"},
			Temperature: []float64{99.9},
		},
	}

	fakeWeatherRepo := newFakeWeatherRepository()

	cacheKey := "cached-forecast"
	if err := fakeWeatherRepo.Set(context.Background(), cacheKey, &cached, time.Hour); err != nil {
		t.Fatal(err)
	}

	if err := fakeWeatherRepo.AddGeoData(context.Background(), "forecast", cacheKey, cachedLocation.Longitude, cachedLocation.Latitude); err != nil {
		t.Fatal(err)
	}

	provider := newStubWeatherProvider(providerForecast, nil)

	weatherService := service.NewWeatherService(provider, fakeWeatherRepo)

	result, err := weatherService.GetForecast(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(cached, *result); diff != "" {
		t.Fatalf("forecast mismatch (-want +got):\n%s", diff)
	}
}

// tryGetForecastFromRepo cache lookup error is ignored
// empty cache key behavior changes
// cache write error from errgroup is ignored
