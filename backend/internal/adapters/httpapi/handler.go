package httpapi

import "github.com/gin-gonic/gin"

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) RegisterRoutes(router *gin.Engine) {
	router.GET("/weather", h.fetchWeatherData)
}

func (h *Handler) fetchWeatherData(c *gin.Context) {}
