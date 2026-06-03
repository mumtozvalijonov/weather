package openmeteo

import "github.com/mumtozvalijonov/weather/internal/core/domain"

type (
	ForecastDto struct {
		Latitude  float64 `json:"latitude" binding:"required,latitude"`
		Longitude float64 `json:"longitude" binding:"required,longitude"`
		Hourly    Hourly  `json:"hourly" binding:"required"`
	}
	Hourly struct {
		Time        []string  `json:"time" binding:"required,dive,len=72,datetime=2026-06-04T23:00"`
		Temperature []float64 `json:"temperature_2m" binding:"required,dive,len=72,number"`
	}
)

func (obj *ForecastDto) intoDomain() domain.Forecast {
	return domain.Forecast{
		Latitude:  obj.Latitude,
		Longitude: obj.Longitude,
		Hourly: domain.Hourly{
			Time:        obj.Hourly.Time,
			Temperature: obj.Hourly.Temperature,
		},
	}
}
