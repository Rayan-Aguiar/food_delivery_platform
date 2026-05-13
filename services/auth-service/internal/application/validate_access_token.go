package application

import (
	"context"

	apperrors "food_delivery_platform/shared/errors"

	"food_delivery_platform/services/auth-service/internal/domain/ports"
	"food_delivery_platform/services/auth-service/internal/domain/valueobjects"
)

type ValidateAccessTokenInput struct {
	AccessToken string
}

type ValidateAccessTokenOutput struct {
	Claims valueobjects.TokenClaims
}

type ValidateAccessTokenUseCase struct {
	credentials  ports.CredentialRepository
	tokenService ports.TokenService
	clock        ports.Clock
}

func NewValidateAccessTokenUseCase(
	credentials ports.CredentialRepository,
	tokenService ports.TokenService,
	clock ports.Clock,
) *ValidateAccessTokenUseCase {
	return &ValidateAccessTokenUseCase{
		credentials:  credentials,
		tokenService: tokenService,
		clock:        clock,
	}
}

func (uc *ValidateAccessTokenUseCase) Execute(ctx context.Context, input ValidateAccessTokenInput) (ValidateAccessTokenOutput, error) {
	if input.AccessToken == "" {
		return ValidateAccessTokenOutput{}, apperrors.Unauthorized("invalid access token", nil)
	}

	claims, err := uc.tokenService.ValidateAccessToken(ctx, input.AccessToken)
	if err != nil {
		return ValidateAccessTokenOutput{}, apperrors.Unauthorized("invalid access token", err)
	}
	if claims.Subject == "" || claims.IsExpired(uc.clock.Now()) {
		return ValidateAccessTokenOutput{}, apperrors.Unauthorized("invalid access token", nil)
	}

	cred, err := uc.credentials.GetByUserID(ctx, claims.Subject)
	if err != nil {
		return ValidateAccessTokenOutput{}, apperrors.Internal("failed to load credential", err)
	}
	if cred == nil || !cred.CanLogin() {
		return ValidateAccessTokenOutput{}, apperrors.Unauthorized("invalid access token", nil)
	}

	return ValidateAccessTokenOutput{Claims: claims}, nil
}
