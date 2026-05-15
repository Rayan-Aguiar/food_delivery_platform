package security_test

import (
	"context"
	"testing"
	"time"

	"food_delivery_platform/services/auth-service/internal/domain/valueobjects"
	"food_delivery_platform/services/auth-service/internal/infrastructure/security"
)

const (
	testJWTSecret      = "01234567890123456789012345678901"
	otherTestJWTSecret = "abcdefghijklmnopqrstuvwxyz123456"
	testJWTIssuer      = "auth-service"
)

func TestNewHMACKeyProvider(t *testing.T) {
	if _, err := security.NewHMACKeyProvider("short-secret"); err == nil {
		t.Fatal("expected error for short secret")
	}

	provider, err := security.NewHMACKeyProvider(testJWTSecret)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if provider == nil {
		t.Fatal("expected provider instance")
	}
}

func TestJWTTokenService_GenerateAndValidateAccessToken(t *testing.T) {
	service, err := security.NewHMACTokenService(testJWTSecret, testJWTIssuer)
	if err != nil {
		t.Fatalf("unexpected constructor error: %v", err)
	}

	now := time.Now().UTC().Truncate(time.Second)
	token, err := service.GenerateAccessToken(context.Background(), valueobjects.TokenClaims{
		Subject:   "user-1",
		IssuedAt:  now,
		ExpiresAt: now.Add(15 * time.Minute),
	})
	if err != nil {
		t.Fatalf("unexpected generate error: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	claims, err := service.ValidateAccessToken(context.Background(), token)
	if err != nil {
		t.Fatalf("unexpected validate error: %v", err)
	}

	if claims.Subject != "user-1" {
		t.Fatalf("expected subject user-1, got %s", claims.Subject)
	}
	if !claims.IssuedAt.Equal(now) {
		t.Fatalf("expected issuedAt %v, got %v", now, claims.IssuedAt)
	}
	if !claims.ExpiresAt.Equal(now.Add(15 * time.Minute)) {
		t.Fatalf("expected expiresAt %v, got %v", now.Add(15*time.Minute), claims.ExpiresAt)
	}
}

func TestJWTTokenService_ValidateAccessTokenRejectsWrongSignature(t *testing.T) {
	validatingService, err := security.NewHMACTokenService(testJWTSecret, testJWTIssuer)
	if err != nil {
		t.Fatalf("unexpected constructor error: %v", err)
	}

	signingService, err := security.NewHMACTokenService(otherTestJWTSecret, testJWTIssuer)
	if err != nil {
		t.Fatalf("unexpected constructor error: %v", err)
	}

	now := time.Now().UTC()
	token, err := signingService.GenerateAccessToken(context.Background(), valueobjects.TokenClaims{
		Subject:   "user-1",
		IssuedAt:  now,
		ExpiresAt: now.Add(15 * time.Minute),
	})
	if err != nil {
		t.Fatalf("unexpected generate error: %v", err)
	}

	if _, err := validatingService.ValidateAccessToken(context.Background(), token); err == nil {
		t.Fatal("expected signature validation error")
	}
}

func TestJWTTokenService_ValidateAccessTokenRejectsExpiredToken(t *testing.T) {
	service, err := security.NewHMACTokenService(testJWTSecret, testJWTIssuer)
	if err != nil {
		t.Fatalf("unexpected constructor error: %v", err)
	}

	now := time.Now().UTC()
	token, err := service.GenerateAccessToken(context.Background(), valueobjects.TokenClaims{
		Subject:   "user-1",
		IssuedAt:  now.Add(-2 * time.Minute),
		ExpiresAt: now.Add(-1 * time.Minute),
	})
	if err != nil {
		t.Fatalf("unexpected generate error: %v", err)
	}

	if _, err := service.ValidateAccessToken(context.Background(), token); err == nil {
		t.Fatal("expected expired token error")
	}
}

func TestJWTTokenService_ValidateAccessTokenRejectsMalformedToken(t *testing.T) {
	service, err := security.NewHMACTokenService(testJWTSecret, testJWTIssuer)
	if err != nil {
		t.Fatalf("unexpected constructor error: %v", err)
	}

	if _, err := service.ValidateAccessToken(context.Background(), "not-a-jwt"); err == nil {
		t.Fatal("expected malformed token error")
	}
}

func TestJWTTokenService_GenerateRefreshToken(t *testing.T) {
	service, err := security.NewHMACTokenService(testJWTSecret, testJWTIssuer)
	if err != nil {
		t.Fatalf("unexpected constructor error: %v", err)
	}

	plain, tokenHash, err := service.GenerateRefreshToken(context.Background())
	if err != nil {
		t.Fatalf("unexpected generate refresh error: %v", err)
	}
	if plain == "" {
		t.Fatal("expected non-empty plain refresh token")
	}
	if tokenHash == "" {
		t.Fatal("expected non-empty refresh token hash")
	}
	if plain == tokenHash {
		t.Fatal("expected hash to differ from plain token")
	}

	// Hash deve ser determinístico.
	recomputedHash, err := service.HashRefreshToken(context.Background(), plain)
	if err != nil {
		t.Fatalf("unexpected hash error: %v", err)
	}
	if recomputedHash != tokenHash {
		t.Fatalf("expected deterministic hash, got %s and %s", recomputedHash, tokenHash)
	}

	// Tokens consecutivos devem ser distintos.
	plain2, _, err := service.GenerateRefreshToken(context.Background())
	if err != nil {
		t.Fatalf("unexpected second generate error: %v", err)
	}
	if plain == plain2 {
		t.Fatal("expected distinct consecutive refresh tokens")
	}
}

func TestJWTTokenService_ContextCanceled(t *testing.T) {
	service, err := security.NewHMACTokenService(testJWTSecret, testJWTIssuer)
	if err != nil {
		t.Fatalf("unexpected constructor error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := service.GenerateAccessToken(ctx, valueobjects.TokenClaims{
		Subject:   "user-1",
		IssuedAt:  time.Now().UTC(),
		ExpiresAt: time.Now().UTC().Add(15 * time.Minute),
	}); err == nil {
		t.Fatal("expected context canceled error in GenerateAccessToken")
	}

	if _, err := service.HashRefreshToken(ctx, "refresh-token"); err == nil {
		t.Fatal("expected context canceled error in HashRefreshToken")
	}
}
