package entities

import (
	"errors"
	"testing"
	"time"

	domainerrors "food_delivery_platform/services/auth-service/internal/domain/errors"
)

func TestNewRefreshSession(t *testing.T) {
	now := time.Now()
	expires := now.Add(10 * time.Minute)

	_, err := NewRefreshSession("", "u1", "hash", expires, "ua", "ip", now)
	if !errors.Is(err, domainerrors.ErrEmptyID) {
		t.Fatalf("expected ErrEmptyID, got %v", err)
	}

	_, err = NewRefreshSession("s1", "", "hash", expires, "ua", "ip", now)
	if !errors.Is(err, domainerrors.ErrEmptyUserID) {
		t.Fatalf("expected ErrEmptyUserID, got %v", err)
	}

	_, err = NewRefreshSession("s1", "u1", "", expires, "ua", "ip", now)
	if !errors.Is(err, domainerrors.ErrEmptyTokenHash) {
		t.Fatalf("expected ErrEmptyTokenHash, got %v", err)
	}

	_, err = NewRefreshSession("s1", "u1", "hash", now, "ua", "ip", now)
	if !errors.Is(err, domainerrors.ErrInvalidTokenTTL) {
		t.Fatalf("expected ErrInvalidTokenTTL, got %v", err)
	}

	s, err := NewRefreshSession("s1", "u1", "hash", expires, "ua", "ip", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.IsRevoked() {
		t.Fatal("session should not be revoked initially")
	}
	if s.IsExpired(now) {
		t.Fatal("session should not be expired initially")
	}
}

func TestRefreshSessionUsageAndRevoke(t *testing.T) {
	now := time.Now()
	s, err := NewRefreshSession("s1", "u1", "hash", now.Add(1*time.Hour), "ua", "ip", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := s.CanBeUsed(now); err != nil {
		t.Fatalf("expected usable session, got %v", err)
	}

	revokeAt := now.Add(1 * time.Minute)
	s.Revoke(revokeAt)
	if !s.IsRevoked() {
		t.Fatal("expected revoked session")
	}
	if s.RevokedAt == nil || !s.RevokedAt.Equal(revokeAt) {
		t.Fatal("expected revoked_at to be set")
	}
	if !errors.Is(s.CanBeUsed(now.Add(2*time.Minute)), domainerrors.ErrRefreshTokenRevoked) {
		t.Fatal("expected ErrRefreshTokenRevoked")
	}
}

func TestRefreshSessionExpirationAndRotate(t *testing.T) {
	now := time.Now()
	expired, err := NewRefreshSession("s1", "u1", "hash", now.Add(1*time.Second), "ua", "ip", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !errors.Is(expired.CanBeUsed(now.Add(2*time.Second)), domainerrors.ErrRefreshTokenExpired) {
		t.Fatal("expected ErrRefreshTokenExpired")
	}

	s, err := NewRefreshSession("s2", "u1", "hash2", now.Add(1*time.Hour), "ua", "ip", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = s.Rotate("", "newhash", now.Add(2*time.Hour), now.Add(1*time.Minute))
	if !errors.Is(err, domainerrors.ErrEmptyID) {
		t.Fatalf("expected ErrEmptyID, got %v", err)
	}

	_, err = s.Rotate("s3", "", now.Add(2*time.Hour), now.Add(1*time.Minute))
	if !errors.Is(err, domainerrors.ErrEmptyTokenHash) {
		t.Fatalf("expected ErrEmptyTokenHash, got %v", err)
	}

	_, err = s.Rotate("s3", "newhash", now, now.Add(1*time.Minute))
	if !errors.Is(err, domainerrors.ErrInvalidTokenTTL) {
		t.Fatalf("expected ErrInvalidTokenTTL, got %v", err)
	}

	rotated, err := s.Rotate("s3", "newhash", now.Add(2*time.Hour), now.Add(1*time.Minute))
	if err != nil {
		t.Fatalf("unexpected rotate error: %v", err)
	}
	if !s.IsRevoked() {
		t.Fatal("old session should be revoked after rotate")
	}
	if rotated.RotatedFromSessionID == nil || *rotated.RotatedFromSessionID != "s2" {
		t.Fatal("expected rotated_from_session_id from old session")
	}
	if rotated.ID != "s3" || rotated.TokenHash != "newhash" {
		t.Fatal("unexpected rotated session values")
	}
}
