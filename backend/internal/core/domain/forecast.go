package domain

type ForecastRequest struct {
	Latitude  float64
	Longitude float64
}

type (
	Forecast struct {
		Latitude  float64
		Longitude float64
		Hourly    Hourly
	}
	Hourly struct {
		Time        []string
		Temperature []float64
	}
)
