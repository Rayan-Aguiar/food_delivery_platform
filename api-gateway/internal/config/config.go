package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	ServiceName    string
	ServiceVersion string
	HTTPPort       string
	LogLevel       string
	RequestTimeout time.Duration
	AuthServiceURL string
}

func Load() Config {
	_ = godotenv.Load(".env")

	timeoutSeconds := getEnvInt("REQUEST_TIMEOUT_SECONDS", 10)

	return Config{
		ServiceName:    getEnv("SERVICE_NAME", "api-gateway"),
		ServiceVersion: getEnv("SERVICE_VERSION", "dev"),
		HTTPPort:       getEnv("HTTP_PORT", "8080"),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		RequestTimeout: time.Duration(timeoutSeconds) * time.Second,
		AuthServiceURL: getEnv("AUTH_SERVICE_URL", "http://localhost:8081"),
	}

}

func getEnv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
