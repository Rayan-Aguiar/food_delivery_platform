package entities

import (
	"errors"
	"testing"
	"time"

	domainerrors "food_delivery_platform/services/auth-service/internal/domain/errors"
	"food_delivery_platform/services/auth-service/internal/domain/valueobjects"
)

func mustEmail(t *testing.T, raw string) valueobjects.Email {
	t.Helper()
	e, err := valueobjects.NewEmail(raw)
	if err != nil {
		t.Fatalf("unexpected email error: %v", err)
	}
	return e
}

func TestNewCredential(t *testing.T) {
	now := time.Now()
	email := mustEmail(t, "user@example.com")

	_, err := NewCredential("", "u1", email, "hash", now)
	if !errors.Is(err, domainerrors.ErrEmptyID) {
		t.Fatalf("expected ErrEmptyID, got: %v", err)
	}

	_, err = NewCredential("c1", "", email, "hash", now)
	if !errors.Is(err, domainerrors.ErrEmptyUserID) {
		t.Fatalf("expected ErrEmptyUserID, got: %v", err)
	}

	_, err = NewCredential("c1", "u1", email, "", now)
	if !errors.Is(err, domainerrors.ErrEmptyPasswordHash) {
		t.Fatalf("expected ErrEmptyPasswordHash, got: %v", err)
	}

	c, err := NewCredential("c1", "u1", email, "hash", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Status != valueobjects.CredentialStatusActive {
		t.Fatalf("unexpected status: %s", c.Status)
	}
	if c.FailedLoginAttempts != 0 {
		t.Fatalf("unexpected failed attempts: %d", c.FailedLoginAttempts)
	}
	if !c.CanLogin() {
		t.Fatal("active credential should be able to login")
	}
}

func TestCredentialStateChanges(t *testing.T) {
	now := time.Now()
	email := mustEmail(t, "user@example.com")
	c, err := NewCredential("c1", "u1", email, "hash", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t2 := now.Add(1 * time.Minute)
	c.RegisterFailedAttempt(t2)
	if c.FailedLoginAttempts != 1 {
		t.Fatalf("expected failed attempts 1, got %d", c.FailedLoginAttempts)
	}
	if !c.UpdatedAt.Equal(t2) {
		t.Fatal("expected updated_at to be changed")
	}

	t3 := now.Add(2 * time.Minute)
	if err := c.RegisterSuccessfulLogin(t3); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.FailedLoginAttempts != 0 {
		t.Fatalf("expected failed attempts reset, got %d", c.FailedLoginAttempts)
	}
	if c.LastLoginAt == nil || !c.LastLoginAt.Equal(t3) {
		t.Fatal("expected last login set")
	}

	t4 := now.Add(3 * time.Minute)
	if err := c.ChangePassword("newhash", t4); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.PasswordHash != "newhash" {
		t.Fatal("expected password hash update")
	}

	t5 := now.Add(4 * time.Minute)
	c.Disable(t5)
	if c.Status != valueobjects.CredentialStatusDisabled {
		t.Fatalf("expected disabled status, got %s", c.Status)
	}
	if c.CanLogin() {
		t.Fatal("disabled credential should not login")
	}

	err = c.RegisterSuccessfulLogin(now.Add(5 * time.Minute))
	if !errors.Is(err, domainerrors.ErrCredentialDisabled) {
		t.Fatalf("expected ErrCredentialDisabled, got %v", err)
	}

	err = c.ChangePassword("", now.Add(6*time.Minute))
	if !errors.Is(err, domainerrors.ErrEmptyPasswordHash) {
		t.Fatalf("expected ErrEmptyPasswordHash, got %v", err)
	}
}
