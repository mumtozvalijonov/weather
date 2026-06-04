package openmeteo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/mumtozvalijonov/weather/internal/core/domain"
)

type Client struct {
	apiClient *http.Client
	baseUrl   url.URL
}

func NewClient(client *http.Client, baseUrl string) (*Client, error) {
	parsedUrl, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}
	return &Client{apiClient: client, baseUrl: *parsedUrl}, nil
}

func (c *Client) GetForecast(ctx context.Context, forecastRequest domain.ForecastRequest) (*domain.Forecast, error) {
	var dto ForecastDto

	u := c.baseUrl
	q := u.Query()
	q.Set("latitude", strconv.FormatFloat(float64(forecastRequest.Latitude), 'f', -1, 32))
	q.Set("longitude", strconv.FormatFloat(float64(forecastRequest.Longitude), 'f', -1, 32))
	q.Set("hourly", "temperature_2m")
	q.Set("forecast_days", "3")
	q.Set("temperature_unit", "celsius")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		u.String(),
		nil,
	)
	if err != nil {
		return nil, err
	}

	resp, err := c.apiClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("third-party API returned an error")
	}

	if err := json.NewDecoder(resp.Body).Decode(&dto); err != nil {
		return nil, err
	}

	return dto.intoDomain(), nil
}
