package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mumtozvalijonov/weather/internal/adapter/httpapi"
	"github.com/mumtozvalijonov/weather/internal/adapter/httpapi/middleware"
	"github.com/mumtozvalijonov/weather/internal/adapter/openmeteo"
	"github.com/mumtozvalijonov/weather/internal/adapter/storage/redis"
	"github.com/mumtozvalijonov/weather/internal/config"
	"github.com/mumtozvalijonov/weather/internal/core/port"
	"github.com/mumtozvalijonov/weather/internal/core/service"
)

func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	openmeteoClient, err := openmeteo.NewClient(&http.Client{Timeout: 10 * time.Second}, cfg.OpenMeteo)
	if err != nil {
		log.Fatalf("failed to initialize openmeteo client: %v", err)
	}

	ctx := context.Background()
	_redis, err := redis.New(ctx, cfg.Redis)
	if err != nil {
		log.Fatalf("failed to connect to Redis: %v", err)
	}
	defer _redis.Close()

	var redisRepo port.WeatherRepository = _redis

	weatherService := service.NewWeatherService(
		openmeteoClient,
		redisRepo,
	)
	handler := httpapi.NewHandler(weatherService)

	router := gin.Default()
	router.Use(middleware.CORSMiddleware(cfg.HTTP.CORSAllowedOrigins))

	weatherRoutes := router.Group("/weather")
	weatherRoutes.Use(middleware.RateLimiterMiddleware())

	handler.RegisterRoutes(weatherRoutes)

	server := &http.Server{
		Addr:              cfg.HTTP.Addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("starting server on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	shutdownCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	<-shutdownCtx.Done()
	stop()

	log.Println("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown failed: %v", err)
	}

	log.Println("server stopped")
}
