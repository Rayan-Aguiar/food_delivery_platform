package application

import (
	"context"
	"errors"

	domainerrors "food_delivery_platform/services/auth-service/internal/domain/errors"
	apperrors "food_delivery_platform/shared/errors"

	"food_delivery_platform/services/auth-service/internal/domain/entities"
	"food_delivery_platform/services/auth-service/internal/domain/ports"
	"food_delivery_platform/services/auth-service/internal/domain/valueobjects"
)

type RegisterUserInput struct {
	Email     string
	Password  string
	UserAgent string
	IPAddress string
}

type RegisterUserOutput struct {
	UserID       string
	CredentialID string
	Tokens       valueobjects.AuthTokens
}

type RegisterUserUseCase struct {
	credentials    ports.CredentialRepository
	sessions       ports.RefreshTokenRepository
	hasher         ports.PasswordHasher
	tokenService   ports.TokenService
	clock          ports.Clock
	idGen          ports.IDGenerator
	ttl            valueobjects.TokenTTL
	passwordPolicy valueobjects.PasswordPolicy
}

func NewRegisterUserUseCase(
	credentials ports.CredentialRepository,
	sessions ports.RefreshTokenRepository,
	hasher ports.PasswordHasher,
	tokenService ports.TokenService,
	clock ports.Clock,
	idGen ports.IDGenerator,
	ttl valueobjects.TokenTTL,
) *RegisterUserUseCase {
	return &RegisterUserUseCase{
		credentials:    credentials,
		sessions:       sessions,
		hasher:         hasher,
		tokenService:   tokenService,
		clock:          clock,
		idGen:          idGen,
		ttl:            ttl,
		passwordPolicy: valueobjects.NewDefaultPasswordPolicy(),
	}
}

func (uc *RegisterUserUseCase) Execute(ctx context.Context, input RegisterUserInput) (RegisterUserOutput, error) {
	email, err := valueobjects.NewEmail(input.Email)
	if err != nil {
		return RegisterUserOutput{}, apperrors.InvalidArgument("invalid email", err)
	}

	if err := uc.passwordPolicy.Validate(input.Password); err != nil {
		return RegisterUserOutput{}, apperrors.InvalidArgument("password does not meet requirements", err)
	}

	existing, err := uc.credentials.GetByEmail(ctx, email.String())
	if err != nil {
		return RegisterUserOutput{}, apperrors.Internal("failed to check email uniqueness", err)
	}
	if existing != nil {
		return RegisterUserOutput{}, apperrors.Conflict("email already registered", nil)
	}

	hash, err := uc.hasher.Hash(ctx, input.Password)
	if err != nil {
		return RegisterUserOutput{}, apperrors.Internal("failed to hash password", err)
	}

	now := uc.clock.Now()
	credentialID := uc.idGen.NewID()
	userID := uc.idGen.NewID()

	cred, err := entities.NewCredential(credentialID, userID, email, hash, now)
	if err != nil {
		return RegisterUserOutput{}, apperrors.Internal("failed to build credential", err)
	}

	if err := uc.credentials.Create(ctx, cred); err != nil {
		if errors.Is(err, domainerrors.ErrEmailAlreadyRegistered) {
			return RegisterUserOutput{}, apperrors.Conflict("email already registered", nil)
		}
		return RegisterUserOutput{}, apperrors.Internal("failed to persist credential", err)
	}

	// Criar sessão inicial — o usuário já sai logado após o registro.
	claims := valueobjects.TokenClaims{
		Subject:   userID,
		IssuedAt:  now,
		ExpiresAt: now.Add(uc.ttl.AccessTTL),
	}
	accessToken, err := uc.tokenService.GenerateAccessToken(ctx, claims)
	if err != nil {
		return RegisterUserOutput{}, apperrors.Internal("failed to generate access token", err)
	}

	plainRefresh, refreshHash, err := uc.tokenService.GenerateRefreshToken(ctx)
	if err != nil {
		return RegisterUserOutput{}, apperrors.Internal("failed to generate refresh token", err)
	}

	session, err := entities.NewRefreshSession(
		uc.idGen.NewID(),
		userID,
		refreshHash,
		now.Add(uc.ttl.RefreshTTL),
		input.UserAgent,
		input.IPAddress,
		now,
	)
	if err != nil {
		return RegisterUserOutput{}, apperrors.Internal("failed to build session", err)
	}

	if err := uc.sessions.Create(ctx, session); err != nil {
		return RegisterUserOutput{}, apperrors.Internal("failed to persist session", err)
	}

	return RegisterUserOutput{
		UserID:       userID,
		CredentialID: credentialID,
		Tokens: valueobjects.AuthTokens{
			AccessToken:  accessToken,
			RefreshToken: plainRefresh,
			TokenType:    "Bearer",
			ExpiresIn:    uc.ttl.AccessSeconds(),
		},
	}, nil
}
