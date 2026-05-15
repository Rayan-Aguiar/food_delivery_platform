package security

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"food_delivery_platform/services/auth-service/internal/domain/ports"
)

type BcryptPasswordHasher struct {
	cost int
}

func NewBcryptPasswordHasher(cost int) (*BcryptPasswordHasher, error) {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		return nil, fmt.Errorf("invalid bcrypt cost: %d", cost)
	}
	return &BcryptPasswordHasher{cost: cost}, nil
}

func (h *BcryptPasswordHasher) Hash(ctx context.Context, plain string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	if plain == "" {
		return "", errors.New("password is required")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(plain), h.cost)
	if err != nil {
		return "", fmt.Errorf("bcrypt hash: %w", err)
	}
	return string(hash), nil
}

func (h *BcryptPasswordHasher) Compare(ctx context.Context, plain, hash string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if plain == "" || hash == "" {
		return errors.New("plain and hash are required")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)); err != nil {
		return fmt.Errorf("bcrypt compare: %w", err)
	}
	return nil
}

var _ ports.PasswordHasher = (*BcryptPasswordHasher)(nil)
