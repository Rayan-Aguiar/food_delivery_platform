package application

import (
	"context"

	apperrors "food_delivery_platform/shared/errors"

	"food_delivery_platform/services/auth-service/internal/domain/ports"
	"food_delivery_platform/services/auth-service/internal/domain/valueobjects"
)

type RefreshAccessTokenInput struct {
	RefreshToken string
}

type RefreshAccessTokenOutput struct {
	UserID string
	Tokens valueobjects.AuthTokens
}

type RefreshAccessTokenUseCase struct {
	credentials  ports.CredentialRepository
	sessions     ports.RefreshTokenRepository
	tokenService ports.TokenService
	clock        ports.Clock
	idGen        ports.IDGenerator
	ttl          valueobjects.TokenTTL
}

func NewRefreshAccessTokenUseCase(
	credentials ports.CredentialRepository,
	sessions ports.RefreshTokenRepository,
	tokenService ports.TokenService,
	clock ports.Clock,
	idGen ports.IDGenerator,
	ttl valueobjects.TokenTTL,
) *RefreshAccessTokenUseCase {
	return &RefreshAccessTokenUseCase{
		credentials:  credentials,
		sessions:     sessions,
		tokenService: tokenService,
		clock:        clock,
		idGen:        idGen,
		ttl:          ttl,
	}
}

func (uc *RefreshAccessTokenUseCase) Execute(ctx context.Context, input RefreshAccessTokenInput) (RefreshAccessTokenOutput, error) {
	if input.RefreshToken == "" {
		return RefreshAccessTokenOutput{}, apperrors.Unauthorized("invalid refresh token", nil)
	}

	tokenHash, err := uc.tokenService.HashRefreshToken(ctx, input.RefreshToken)
	if err != nil {
		return RefreshAccessTokenOutput{}, apperrors.Internal("failed to hash refresh token", err)
	}

	session, err := uc.sessions.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return RefreshAccessTokenOutput{}, apperrors.Internal("failed to load session", err)
	}
	if session == nil {
		return RefreshAccessTokenOutput{}, apperrors.Unauthorized("invalid refresh token", nil)
	}

	now := uc.clock.Now()
	if err := session.CanBeUsed(now); err != nil {
		return RefreshAccessTokenOutput{}, apperrors.Unauthorized("invalid refresh token", err)
	}

	cred, err := uc.credentials.GetByUserID(ctx, session.UserID)
	if err != nil {
		return RefreshAccessTokenOutput{}, apperrors.Internal("failed to load credential", err)
	}
	if cred == nil || !cred.CanLogin() {
		return RefreshAccessTokenOutput{}, apperrors.Unauthorized("invalid refresh token", nil)
	}

	accessClaims := valueobjects.TokenClaims{
		Subject:   session.UserID,
		IssuedAt:  now,
		ExpiresAt: now.Add(uc.ttl.AccessTTL),
	}
	accessToken, err := uc.tokenService.GenerateAccessToken(ctx, accessClaims)
	if err != nil {
		return RefreshAccessTokenOutput{}, apperrors.Internal("failed to generate access token", err)
	}

	newPlain, newHash, err := uc.tokenService.GenerateRefreshToken(ctx)
	if err != nil {
		return RefreshAccessTokenOutput{}, apperrors.Internal("failed to generate refresh token", err)
	}

	rotated, err := session.Rotate(uc.idGen.NewID(), newHash, now.Add(uc.ttl.RefreshTTL), now)
	if err != nil {
		return RefreshAccessTokenOutput{}, apperrors.Unauthorized("invalid refresh token", err)
	}

	if err := uc.sessions.Update(ctx, session); err != nil {
		return RefreshAccessTokenOutput{}, apperrors.Internal("failed to revoke previous session", err)
	}
	if err := uc.sessions.Create(ctx, rotated); err != nil {
		return RefreshAccessTokenOutput{}, apperrors.Internal("failed to persist rotated session", err)
	}

	return RefreshAccessTokenOutput{
		UserID: session.UserID,
		Tokens: valueobjects.AuthTokens{
			AccessToken:  accessToken,
			RefreshToken: newPlain,
			TokenType:    "Bearer",
			ExpiresIn:    uc.ttl.AccessSeconds(),
		},
	}, nil
}
