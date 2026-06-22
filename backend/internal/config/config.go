package config

import (
	"fmt"
	"os"

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
		Addr string
	}
)

func Load() (Config, error) {
	_ = godotenv.Load(".env")

	httpAddr, err := requiredEnv("HTTP_ADDR")
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
			Addr: httpAddr,
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

func requiredEnv(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("%s is not configured", key)
	}

	return value, nil
}
