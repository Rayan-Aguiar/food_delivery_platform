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
	RabbitMQURL    string
	BcryptCost     int
	JWTSecret      string
	JWTIssuer      string
	JWTAccessTTL   time.Duration
	JWTRefreshTTL  time.Duration
	OTelEnabled    bool
}

func Load() Config {
	// Carrega .env local quando existir. Em producao, variaveis vem do ambiente.
	_ = godotenv.Load(".env")

	cfg := Config{
		ServiceName:    getEnv("SERVICE_NAME", "auth-service"),
		ServiceVersion: getEnv("SERVICE_VERSION", "dev"),
		HTTPPort:       getEnv("HTTP_PORT", "8081"),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		MongoURI:       getEnv("MONGO_URI", ""),
		MongoDBName:    getEnv("MONGO_DB_NAME", "auth_db"),
		RabbitMQURL:    getEnv("RABBITMQ_URL", ""),
		BcryptCost:     getEnvInt("BCRYPT_COST", 12),
		JWTSecret:      getEnv("JWT_SECRET", ""),
		JWTIssuer:      getEnv("JWT_ISSUER", "auth-service"),
		JWTAccessTTL:   time.Duration(getEnvInt("JWT_ACCESS_TTL", 15)) * time.Minute,
		JWTRefreshTTL:  time.Duration(getEnvInt("JWT_REFRESH_TTL", 60)) * time.Minute,
		OTelEnabled:    getEnvBool("OTEL_ENABLED", false),
	}
	// Faixa valida do bcrypt: 4..31. Em valor invalido, usa fallback seguro.
	if cfg.BcryptCost < 4 || cfg.BcryptCost > 31 {
		cfg.BcryptCost = 12
	}

	if cfg.ServiceName != "" && os.Getenv("JWT_ISSUER") == "" {
		cfg.JWTIssuer = cfg.ServiceName
	}

	if cfg.BcryptCost < 4 || cfg.BcryptCost > 31 {
		cfg.BcryptCost = 12
	}

	if cfg.JWTAccessTTL <= 0 {
		cfg.JWTAccessTTL = 15 * time.Minute
	}
	if cfg.JWTRefreshTTL <= cfg.JWTAccessTTL {
		cfg.JWTRefreshTTL = 7 * 24 * time.Hour
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

func getEnvBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}
