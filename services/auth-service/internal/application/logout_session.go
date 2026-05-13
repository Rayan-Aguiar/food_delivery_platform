package application

import (
	"context"

	apperrors "food_delivery_platform/shared/errors"

	"food_delivery_platform/services/auth-service/internal/domain/ports"
)

type LogoutSessionInput struct {
	RefreshToken string
}

type LogoutSessionUseCase struct {
	sessions     ports.RefreshTokenRepository
	tokenService ports.TokenService
}

func NewLogoutSessionUseCase(
	sessions ports.RefreshTokenRepository,
	tokenService ports.TokenService,
) *LogoutSessionUseCase {
	return &LogoutSessionUseCase{sessions: sessions, tokenService: tokenService}
}

func (uc *LogoutSessionUseCase) Execute(ctx context.Context, input LogoutSessionInput) error {
	if input.RefreshToken == "" {
		return apperrors.Unauthorized("invalid refresh token", nil)
	}

	tokenHash, err := uc.tokenService.HashRefreshToken(ctx, input.RefreshToken)
	if err != nil {
		return apperrors.Internal("failed to hash refresh token", err)
	}

	session, err := uc.sessions.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return apperrors.Internal("failed to load session", err)
	}
	if session == nil {
		return apperrors.Unauthorized("invalid refresh token", nil)
	}

	if session.IsRevoked() {
		return nil
	}

	if err := uc.sessions.Revoke(ctx, session.ID); err != nil {
		return apperrors.Internal("failed to revoke session", err)
	}

	return nil
}
