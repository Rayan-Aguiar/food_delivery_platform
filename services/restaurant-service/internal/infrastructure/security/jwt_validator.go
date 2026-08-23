package security

import (
	"context"
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

// JWTValidator verifica tokens HS256 emitidos pelo auth-service. O
// restaurant-service nunca emite token, apenas confia no mesmo segredo
// compartilhado para validar quem ja se autenticou la.
type JWTValidator struct {
	secret []byte
	issuer string
}

func NewJWTValidator(secret, issuer string) (*JWTValidator, error) {
	if len(secret) < 32 {
		return nil, errors.New("jwt secret must have at least 32 characters")
	}
	if issuer == "" {
		return nil, errors.New("jwt issuer is required")
	}

	return &JWTValidator{secret: []byte(secret), issuer: issuer}, nil
}

func (v *JWTValidator) ValidateAccessToken(ctx context.Context, tokenString string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	if tokenString == "" {
		return "", errors.New("token is required")
	}

	claims := &jwt.RegisteredClaims{}
	parsed, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		return v.secret, nil
	},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(v.issuer),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return "", fmt.Errorf("parse access token: %w", err)
	}
	if !parsed.Valid || claims.Subject == "" {
		return "", errors.New("invalid access token claims")
	}

	return claims.Subject, nil
}
