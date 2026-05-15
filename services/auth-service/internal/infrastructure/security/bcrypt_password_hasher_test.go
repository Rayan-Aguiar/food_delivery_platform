package security

import (
	"context"
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestNewBcryptPasswordHasher(t *testing.T) {
	_, err := NewBcryptPasswordHasher(3)
	if err == nil {
		t.Fatal("expected error for invalid cost")
	}

	h, err := NewBcryptPasswordHasher(12)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h == nil {
		t.Fatal("expected hasher instance")
	}
}

func TestBcryptPasswordHasher_HashAndCompare(t *testing.T) {
	h, err := NewBcryptPasswordHasher(12)
	if err != nil {
		t.Fatalf("unexpected constructor error: %v", err)
	}

	hash, err := h.Hash(context.Background(), "Secret1@")
	if err != nil {
		t.Fatalf("unexpected hash error: %v", err)
	}
	if hash == "" {
		t.Fatal("expected non-empty hash")
	}

	cost, err := bcrypt.Cost([]byte(hash))
	if err != nil {
		t.Fatalf("unexpected cost parse error: %v", err)
	}
	if cost != 12 {
		t.Fatalf("expected cost 12, got %d", cost)
	}

	if err := h.Compare(context.Background(), "Secret1@", hash); err != nil {
		t.Fatalf("expected compare success, got: %v", err)
	}
}

func TestBcryptPasswordHasher_CompareWrongPassword(t *testing.T) {
	h, _ := NewBcryptPasswordHasher(10)
	hash, _ := h.Hash(context.Background(), "Secret1@")

	err := h.Compare(context.Background(), "Wrong1@", hash)
	if err == nil {
		t.Fatal("expected compare error")
	}
	if !errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		t.Fatalf("expected mismatched password error, got: %v", err)
	}
}

func TestBcryptPasswordHasher_ContextCanceled(t *testing.T) {
	h, _ := NewBcryptPasswordHasher(10)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := h.Hash(ctx, "Secret1@"); err == nil {
		t.Fatal("expected context cancellation error in Hash")
	}
	if err := h.Compare(ctx, "Secret1@", "$2a$10$invalidinvalidinvalidinvalidinvalidinvalidinv.q5m5a"); err == nil {
		t.Fatal("expected context cancellation error in Compare")
	}
}
