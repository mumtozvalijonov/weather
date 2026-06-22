package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/mumtozvalijonov/weather/internal/adapter/openmeteo"
	"github.com/mumtozvalijonov/weather/internal/adapter/storage/redis"
)

type (
	Config struct {
		HTTP      HTTP
		Redis     redis.Config
		OpenMeteo openmeteo.Config
	}
	HTTP struct {
		Addr               string
		CORSAllowedOrigins []string
	}
)

func Load() (Config, error) {
	_ = godotenv.Load(".env")

	httpAddr, err := requiredEnv("HTTP_ADDR")
	if err != nil {
		return Config{}, err
	}

	corsAllowedOrigins, err := requiredEnv("CORS_ALLOWED_ORIGINS")
	if err != nil {
		return Config{}, err
	}

	redisAddr, err := requiredEnv("REDIS_ADDR")
	if err != nil {
		return Config{}, err
	}

	openMeteoBaseURL, err := requiredEnv("OPEN_METEO_BASE_URL")
	if err != nil {
		return Config{}, err
	}

	return Config{
		HTTP: HTTP{
			Addr:               httpAddr,
			CORSAllowedOrigins: splitCSV(corsAllowedOrigins),
		},
		Redis: redis.Config{
			Addr:     redisAddr,
			Password: os.Getenv("REDIS_PASSWORD"),
		},
		OpenMeteo: openmeteo.Config{
			BaseURL: openMeteoBaseURL,
		},
	}, nil
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func requiredEnv(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("%s is not configured", key)
	}

	return value, nil
}
