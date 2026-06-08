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
	MongoURI       string
	MongoDBName    string
}

func Load() Config {
	_ = godotenv.Load(".env")

	timeoutSeconds := getEnvInt("REQUEST_TIMEOUT_SECONDS", 10)

	return Config{
		ServiceName:    getEnv("SERVICE_NAME", "restaurant-service"),
		ServiceVersion: getEnv("SERVICE_VERSION", "dev"),
		HTTPPort:       getEnv("HTTP_PORT", "8083"),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		RequestTimeout: time.Duration(timeoutSeconds) * time.Second,
		MongoURI:       getEnv("MONGO_URI", ""),
		MongoDBName:    getEnv("MONGO_DB_NAME", "restaurant_db"),
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
