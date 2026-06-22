package httpapi

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mumtozvalijonov/weather/internal/core/domain"
	"github.com/mumtozvalijonov/weather/internal/core/port"
)

type WeatherApiResponse struct {
	GenerationTimeMS float64 `json:"generationtime_ms"`
}
type Handler struct {
	svc port.WeatherService
}

func NewHandler(svc port.WeatherService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(router gin.IRoutes) {
	router.GET("", h.fetchWeatherData)
}

type fetchWeatherDataRequest struct {
	Latitude  float64 `form:"lat" binding:"required,latitude"`
	Longitude float64 `form:"lon" binding:"required,longitude"`
}

func (h *Handler) fetchWeatherData(ctx *gin.Context) {
	var req fetchWeatherDataRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	forecast, err := h.svc.GetForecast(ctx.Request.Context(), domain.ForecastRequest{Latitude: req.Latitude, Longitude: req.Longitude})
	if err != nil {
		log.Printf("failed to get forecast: %+v", err)
		ctx.JSON(http.StatusBadGateway, gin.H{"error": "failed to get forecast"})
		return
	}

	ctx.JSON(http.StatusOK, forecast)
}
