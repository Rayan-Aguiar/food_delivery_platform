package security

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "this-is-a-test-secret-with-32-chars"

func signToken(t *testing.T, secret, issuer, subject string, issuedAt, expiresAt time.Time) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   subject,
		Issuer:    issuer,
		IssuedAt:  jwt.NewNumericDate(issuedAt),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
	})
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return signed
}

func TestNewJWTValidator(t *testing.T) {
	if _, err := NewJWTValidator("too-short", "auth-service"); err == nil {
		t.Error("expected error for short secret")
	}
	if _, err := NewJWTValidator(testSecret, ""); err == nil {
		t.Error("expected error for empty issuer")
	}
	if _, err := NewJWTValidator(testSecret, "auth-service"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestJWTValidator_ValidateAccessToken(t *testing.T) {
	v, err := NewJWTValidator(testSecret, "auth-service")
	if err != nil {
		t.Fatalf("unexpected constructor error: %v", err)
	}
	ctx := context.Background()
	now := time.Now()

	t.Run("valid token returns subject", func(t *testing.T) {
		token := signToken(t, testSecret, "auth-service", "user-1", now, now.Add(time.Hour))
		subject, err := v.ValidateAccessToken(ctx, token)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if subject != "user-1" {
			t.Errorf("subject = %q, want user-1", subject)
		}
	})

	t.Run("empty token", func(t *testing.T) {
		if _, err := v.ValidateAccessToken(ctx, ""); err == nil {
			t.Error("expected error for empty token")
		}
	})

	t.Run("expired token", func(t *testing.T) {
		token := signToken(t, testSecret, "auth-service", "user-1", now.Add(-2*time.Hour), now.Add(-time.Hour))
		if _, err := v.ValidateAccessToken(ctx, token); err == nil {
			t.Error("expected error for expired token")
		}
	})

	t.Run("wrong issuer", func(t *testing.T) {
		token := signToken(t, testSecret, "someone-else", "user-1", now, now.Add(time.Hour))
		if _, err := v.ValidateAccessToken(ctx, token); err == nil {
			t.Error("expected error for wrong issuer")
		}
	})

	t.Run("wrong secret", func(t *testing.T) {
		token := signToken(t, "another-secret-with-32-characters!!", "auth-service", "user-1", now, now.Add(time.Hour))
		if _, err := v.ValidateAccessToken(ctx, token); err == nil {
			t.Error("expected error for wrong secret")
		}
	})

	t.Run("missing subject", func(t *testing.T) {
		token := signToken(t, testSecret, "auth-service", "", now, now.Add(time.Hour))
		if _, err := v.ValidateAccessToken(ctx, token); err == nil {
			t.Error("expected error for missing subject")
		}
	})

	t.Run("wrong signing method", func(t *testing.T) {
		token := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.RegisteredClaims{
			Subject:   "user-1",
			Issuer:    "auth-service",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
		})
		signed, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
		if err != nil {
			t.Fatalf("sign none token: %v", err)
		}
		if _, err := v.ValidateAccessToken(ctx, signed); err == nil {
			t.Error("expected error for none-algorithm token")
		}
	})
}
