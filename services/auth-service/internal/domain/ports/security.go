package ports

import (
	"context"

	"food_delivery_platform/services/auth-service/internal/domain/valueobjects"
)

type PasswordHasher interface {
	Hash(ctx context.Context, plain string) (string, error)
	Compare(ctx context.Context, plain, hash string) error
}

type TokenService interface {
	GenerateAccessToken(ctx context.Context, claims valueobjects.TokenClaims) (string, error)
	ValidateAccessToken(ctx context.Context, token string) (valueobjects.TokenClaims, error)

	GenerateRefreshToken(ctx context.Context) (plain string, tokenHash string, err error)
	HashRefreshToken(ctx context.Context, plain string) (string, error)
}
