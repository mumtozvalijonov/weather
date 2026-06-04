package service_test

import (
	"context"

	"github.com/mumtozvalijonov/weather/internal/core/domain"
)

type stubWeatherProvider struct {
	forecast domain.Forecast
	err      error
}

func newStubWeatherProvider(forecast domain.Forecast, err error) *stubWeatherProvider {
	return &stubWeatherProvider{forecast: forecast, err: err}
}

func (f *stubWeatherProvider) GetForecast(ctx context.Context, req domain.ForecastRequest) (*domain.Forecast, error) {
	return &f.forecast, f.err
}
