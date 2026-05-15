package security

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"

	"github.com/golang-jwt/jwt/v5"

	"food_delivery_platform/services/auth-service/internal/domain/ports"
	"food_delivery_platform/services/auth-service/internal/domain/valueobjects"
)

type JWTKeyProvider interface {
	SigningMethod() jwt.SigningMethod
	SignKey(ctx context.Context) (any, error)
	VerifyKey(ctx context.Context, token *jwt.Token) (any, error)
}

type HMACKeyProvider struct {
	secret []byte
}

func NewHMACKeyProvider(secret string) (*HMACKeyProvider, error) {
	if len(secret) < 32 {
		return nil, errors.New("jwt secret must have at least 32 characters")
	}

	return &HMACKeyProvider{secret: []byte(secret)}, nil
}

func (p *HMACKeyProvider) SigningMethod() jwt.SigningMethod {
	return jwt.SigningMethodHS256
}

func (p *HMACKeyProvider) SignKey(ctx context.Context) (any, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if len(p.secret) == 0 {
		return nil, errors.New("jwt secret is required")
	}

	return p.secret, nil
}

func (p *HMACKeyProvider) VerifyKey(ctx context.Context, token *jwt.Token) (any, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if token == nil {
		return nil, errors.New("token is required")
	}
	if token.Method == nil {
		return nil, errors.New("signing method is required")
	}
	if token.Method.Alg() != p.SigningMethod().Alg() {
		return nil, fmt.Errorf("unexpected signing method: %s", token.Method.Alg())
	}

	return p.secret, nil
}

type JWTTokenService struct {
	issuer      string
	keyProvider JWTKeyProvider
	random      io.Reader
}

func NewJWTTokenService(issuer string, keyProvider JWTKeyProvider) (*JWTTokenService, error) {
	if issuer == "" {
		return nil, errors.New("jwt issuer is required")
	}
	if keyProvider == nil {
		return nil, errors.New("jwt key provider is required")
	}

	return &JWTTokenService{
		issuer:      issuer,
		keyProvider: keyProvider,
		random:      rand.Reader,
	}, nil
}

func NewHMACTokenService(secret, issuer string) (*JWTTokenService, error) {
	keyProvider, err := NewHMACKeyProvider(secret)
	if err != nil {
		return nil, err
	}

	return NewJWTTokenService(issuer, keyProvider)
}

func (s *JWTTokenService) GenerateAccessToken(ctx context.Context, claims valueobjects.TokenClaims) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	if claims.Subject == "" {
		return "", errors.New("token subject is required")
	}
	if claims.IssuedAt.IsZero() || claims.ExpiresAt.IsZero() {
		return "", errors.New("token issuedAt and expiresAt are required")
	}
	if !claims.ExpiresAt.After(claims.IssuedAt) {
		return "", errors.New("token expiresAt must be after issuedAt")
	}

	signKey, err := s.keyProvider.SignKey(ctx)
	if err != nil {
		return "", fmt.Errorf("resolve signing key: %w", err)
	}

	token := jwt.NewWithClaims(s.keyProvider.SigningMethod(), jwt.RegisteredClaims{
		Subject:   claims.Subject,
		Issuer:    s.issuer,
		IssuedAt:  jwt.NewNumericDate(claims.IssuedAt.UTC()),
		ExpiresAt: jwt.NewNumericDate(claims.ExpiresAt.UTC()),
	})

	signed, err := token.SignedString(signKey)
	if err != nil {
		return "", fmt.Errorf("sign access token: %w", err)
	}

	return signed, nil
}

func (s *JWTTokenService) ValidateAccessToken(ctx context.Context, tokenString string) (valueobjects.TokenClaims, error) {
	select {
	case <-ctx.Done():
		return valueobjects.TokenClaims{}, ctx.Err()
	default:
	}

	if tokenString == "" {
		return valueobjects.TokenClaims{}, errors.New("token is required")
	}

	parsed, err := jwt.ParseWithClaims(
		tokenString,
		&jwt.RegisteredClaims{},
		func(token *jwt.Token) (any, error) {
			return s.keyProvider.VerifyKey(ctx, token)
		},
		jwt.WithValidMethods([]string{s.keyProvider.SigningMethod().Alg()}),
		jwt.WithIssuer(s.issuer),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return valueobjects.TokenClaims{}, fmt.Errorf("parse access token: %w", err)
	}

	registeredClaims, ok := parsed.Claims.(*jwt.RegisteredClaims)
	if !ok || !parsed.Valid {
		return valueobjects.TokenClaims{}, errors.New("invalid access token claims")
	}
	if registeredClaims.Subject == "" {
		return valueobjects.TokenClaims{}, errors.New("token subject is required")
	}
	if registeredClaims.IssuedAt == nil || registeredClaims.ExpiresAt == nil {
		return valueobjects.TokenClaims{}, errors.New("token iat and exp are required")
	}

	return valueobjects.TokenClaims{
		Subject:   registeredClaims.Subject,
		IssuedAt:  registeredClaims.IssuedAt.Time.UTC(),
		ExpiresAt: registeredClaims.ExpiresAt.Time.UTC(),
	}, nil
}

func (s *JWTTokenService) GenerateRefreshToken(ctx context.Context) (string, string, error) {
	select {
	case <-ctx.Done():
		return "", "", ctx.Err()
	default:
	}

	raw := make([]byte, 32)
	if _, err := io.ReadFull(s.random, raw); err != nil {
		return "", "", fmt.Errorf("read refresh token entropy: %w", err)
	}

	plain := base64.RawURLEncoding.EncodeToString(raw)
	tokenHash, err := s.HashRefreshToken(ctx, plain)
	if err != nil {
		return "", "", err
	}

	return plain, tokenHash, nil
}

func (s *JWTTokenService) HashRefreshToken(ctx context.Context, plain string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	if plain == "" {
		return "", errors.New("refresh token is required")
	}

	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:]), nil
}

var _ ports.TokenService = (*JWTTokenService)(nil)
