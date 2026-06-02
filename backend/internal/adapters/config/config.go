package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type (
	Config struct {
		HTTP *HTTP
	}
	// Redis struct {
	// 	Addr     string
	// 	Password string
	// }
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
	http := HTTP{
		Addr: httpAddr,
	}

	return Config{
		&http,
	}, nil
}

func requiredEnv(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("%s is not configured", key)
	}

	return value, nil
}
