package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	ServiceName    string
	HTTPPort       string
	LogLevel       string
	RequestTimeout time.Duration
	MongoURI       string
	MongoDBName    string
	RabbitMQURL    string
	BcryptCost     int
}

func Load() Config {
	// Carrega .env local quando existir. Em producao, variaveis vem do ambiente.
	_ = godotenv.Load(".env")

	cfg := Config{
		ServiceName: getEnv("SERVICE_NAME", "auth-service"),
		HTTPPort:    getEnv("HTTP_PORT", "8081"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		MongoURI:    getEnv("MONGO_URI", ""),
		MongoDBName: getEnv("MONGO_DB_NAME", "auth_db"),
		RabbitMQURL: getEnv("RABBITMQ_URL", ""),
		BcryptCost:  getEnvInt("BCRYPT_COST", 12),
	}
	// Faixa valida do bcrypt: 4..31. Em valor invalido, usa fallback seguro.
	if cfg.BcryptCost < 4 || cfg.BcryptCost > 31 {
		cfg.BcryptCost = 12
	}

	timeoutSeconds := getEnvInt("REQUEST_TIMEOUT_SECONDS", 10)
	cfg.RequestTimeout = time.Duration(timeoutSeconds) * time.Second

	return cfg
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
