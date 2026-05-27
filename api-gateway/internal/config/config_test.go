package config

import (
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("SERVICE_NAME", "")
	t.Setenv("SERVICE_VERSION", "")
	t.Setenv("HTTP_PORT", "")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("REQUEST_TIMEOUT_SECONDS", "")
	t.Setenv("AUTH_SERVICE_URL", "")

	cfg := Load()

	if cfg.ServiceName != "api-gateway" {
		t.Fatalf("expected default service name api-gateway, got %q", cfg.ServiceName)
	}
	if cfg.ServiceVersion != "dev" {
		t.Fatalf("expected default service version dev, got %q", cfg.ServiceVersion)
	}
	if cfg.HTTPPort != "8080" {
		t.Fatalf("expected default http port 8080, got %q", cfg.HTTPPort)
	}
	if cfg.LogLevel != "info" {
		t.Fatalf("expected default log level info, got %q", cfg.LogLevel)
	}
	if cfg.RequestTimeout != 10*time.Second {
		t.Fatalf("expected default timeout 10s, got %s", cfg.RequestTimeout)
	}
	if cfg.AuthServiceURL != "http://localhost:8081" {
		t.Fatalf("expected default auth url http://localhost:8081, got %q", cfg.AuthServiceURL)
	}
}

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("SERVICE_NAME", "edge-gateway")
	t.Setenv("SERVICE_VERSION", "1.2.3")
	t.Setenv("HTTP_PORT", "9090")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("REQUEST_TIMEOUT_SECONDS", "3")
	t.Setenv("AUTH_SERVICE_URL", "http://auth-service:8081")

	cfg := Load()

	if cfg.ServiceName != "edge-gateway" {
		t.Fatalf("expected service name edge-gateway, got %q", cfg.ServiceName)
	}
	if cfg.ServiceVersion != "1.2.3" {
		t.Fatalf("expected service version 1.2.3, got %q", cfg.ServiceVersion)
	}
	if cfg.HTTPPort != "9090" {
		t.Fatalf("expected port 9090, got %q", cfg.HTTPPort)
	}
	if cfg.LogLevel != "debug" {
		t.Fatalf("expected log level debug, got %q", cfg.LogLevel)
	}
	if cfg.RequestTimeout != 3*time.Second {
		t.Fatalf("expected timeout 3s, got %s", cfg.RequestTimeout)
	}
	if cfg.AuthServiceURL != "http://auth-service:8081" {
		t.Fatalf("expected auth url http://auth-service:8081, got %q", cfg.AuthServiceURL)
	}
}
