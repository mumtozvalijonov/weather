package service_test

import (
	"context"
	"errors"
	"testing"

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

			service := service.NewWeatherService(
				stubWeatherProvider,
				fakeWeatherRepo,
			)

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

		})
	}
}

func TestWeatherService_GetForecast_CacheMiss_PersistsForecastForNearbyRequest(t *testing.T) {
	ctx := context.Background()

	firstReq := domain.ForecastRequest{
		Latitude:  40.7128,
		Longitude: -74.0060,
	}
	nearbyReq := domain.ForecastRequest{
		Latitude:  40.7130,
		Longitude: -74.0062,
	}

	providerForecast := domain.Forecast{
		Latitude:  firstReq.Latitude,
		Longitude: firstReq.Longitude,
		Hourly: domain.Hourly{
			Time:        []string{"2026-06-04T00:00"},
			Temperature: []float64{21.5},
		},
	}
	fallbackProviderForecast := domain.Forecast{
		Latitude:  nearbyReq.Latitude,
		Longitude: nearbyReq.Longitude,
		Hourly: domain.Hourly{
			Time:        []string{"provider-called-again"},
			Temperature: []float64{99.9},
		},
	}

	fakeWeatherRepo := newFakeWeatherRepository()

	weatherService := service.NewWeatherService(
		newStubWeatherProvider(providerForecast, nil),
		fakeWeatherRepo,
	)

	firstResult, err := weatherService.GetForecast(ctx, firstReq)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(providerForecast, *firstResult); diff != "" {
		t.Fatalf("first forecast mismatch (-want +got):\n%s", diff)
	}

	weatherService = service.NewWeatherService(
		newStubWeatherProvider(fallbackProviderForecast, nil),
		fakeWeatherRepo,
	)

	secondResult, err := weatherService.GetForecast(ctx, nearbyReq)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(providerForecast, *secondResult); diff != "" {
		t.Fatalf("second forecast should come from cache (-want +got):\n%s", diff)
	}
}

func TestWeatherService_GetForecast_ReturnsFreshForecastForRequestOutsideCachedArea(t *testing.T) {
	ctx := context.Background()

	originalReq := domain.ForecastRequest{
		Latitude:  40.7128,
		Longitude: -74.0060,
	}
	outsideReq := domain.ForecastRequest{
		Latitude:  40.7260,
		Longitude: -74.0260,
	}

	originalForecast := domain.Forecast{
		Latitude:  originalReq.Latitude,
		Longitude: originalReq.Longitude,
		Hourly: domain.Hourly{
			Time:        []string{"2026-06-04T00:00"},
			Temperature: []float64{21.5},
		},
	}
	freshForecast := domain.Forecast{
		Latitude:  outsideReq.Latitude,
		Longitude: outsideReq.Longitude,
		Hourly: domain.Hourly{
			Time:        []string{"2026-06-04T01:00"},
			Temperature: []float64{19.8},
		},
	}

	fakeWeatherRepo := newFakeWeatherRepository()

	weatherService := service.NewWeatherService(
		newStubWeatherProvider(originalForecast, nil),
		fakeWeatherRepo,
	)

	firstResult, err := weatherService.GetForecast(ctx, originalReq)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(originalForecast, *firstResult); diff != "" {
		t.Fatalf("original forecast mismatch (-want +got):\n%s", diff)
	}

	weatherService = service.NewWeatherService(
		newStubWeatherProvider(freshForecast, nil),
		fakeWeatherRepo,
	)

	secondResult, err := weatherService.GetForecast(ctx, outsideReq)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(freshForecast, *secondResult); diff != "" {
		t.Fatalf("forecast outside cached area mismatch (-want +got):\n%s", diff)
	}
}
