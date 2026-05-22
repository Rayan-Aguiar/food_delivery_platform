package application

import (
	"context"
	"time"

	apperrors "food_delivery_platform/shared/errors"

	"food_delivery_platform/services/auth-service/internal/domain/entities"
	"food_delivery_platform/services/auth-service/internal/domain/ports"
	"food_delivery_platform/services/auth-service/internal/domain/valueobjects"
)

type LoginUserInput struct {
	Email     string
	Password  string
	UserAgent string
	IPAddress string

	CorrelationID  string
	CausationID    string
	Traceparent    string
	IdempotencyKey string
}

type LoginUserOutput struct {
	UserID string
	Tokens valueobjects.AuthTokens
}

type LoginUserUseCase struct {
	credentials  ports.CredentialRepository
	sessions     ports.RefreshTokenRepository
	hasher       ports.PasswordHasher
	tokenService ports.TokenService
	clock        ports.Clock
	idGen        ports.IDGenerator
	ttl          valueobjects.TokenTTL
	events       ports.AuthEventPublisher
}

func NewLoginUserUseCase(
	credentials ports.CredentialRepository,
	sessions ports.RefreshTokenRepository,
	hasher ports.PasswordHasher,
	tokenService ports.TokenService,
	clock ports.Clock,
	idGen ports.IDGenerator,
	ttl valueobjects.TokenTTL,
	eventPublisher ...ports.AuthEventPublisher,
) *LoginUserUseCase {
	publisher := ports.AuthEventPublisher(noopAuthEventPublisher{})
	if len(eventPublisher) > 0 && eventPublisher[0] != nil {
		publisher = eventPublisher[0]
	}

	return &LoginUserUseCase{
		credentials:  credentials,
		sessions:     sessions,
		hasher:       hasher,
		tokenService: tokenService,
		clock:        clock,
		idGen:        idGen,
		ttl:          ttl,
		events:       publisher,
	}
}

func (uc *LoginUserUseCase) Execute(ctx context.Context, input LoginUserInput) (LoginUserOutput, error) {
	email, err := valueobjects.NewEmail(input.Email)
	if err != nil {
		// email inválido não revela se o usuário existe — retorna Unauthorized
		return LoginUserOutput{}, apperrors.Unauthorized("invalid credentials", nil)
	}

	cred, err := uc.credentials.GetByEmail(ctx, email.String())
	if err != nil {
		return LoginUserOutput{}, apperrors.Internal("failed to lookup credential", err)
	}
	if cred == nil || !cred.CanLogin() {
		return LoginUserOutput{}, apperrors.Unauthorized("invalid credentials", nil)
	}

	now := uc.clock.Now()

	if err := uc.hasher.Compare(ctx, input.Password, cred.PasswordHash); err != nil {
		cred.RegisterFailedAttempt(now)
		_ = uc.credentials.Update(ctx, cred) // best-effort; não falha o login por erro de infra
		return LoginUserOutput{}, apperrors.Unauthorized("invalid credentials", nil)
	}

	if err := cred.RegisterSuccessfulLogin(now); err != nil {
		return LoginUserOutput{}, apperrors.Unauthorized("invalid credentials", nil)
	}
	if err := uc.credentials.Update(ctx, cred); err != nil {
		return LoginUserOutput{}, apperrors.Internal("failed to update credential", err)
	}

	claims := valueobjects.TokenClaims{
		Subject:   cred.UserID,
		IssuedAt:  now,
		ExpiresAt: now.Add(uc.ttl.AccessTTL),
	}
	accessToken, err := uc.tokenService.GenerateAccessToken(ctx, claims)
	if err != nil {
		return LoginUserOutput{}, apperrors.Internal("failed to generate access token", err)
	}

	plainRefresh, refreshHash, err := uc.tokenService.GenerateRefreshToken(ctx)
	if err != nil {
		return LoginUserOutput{}, apperrors.Internal("failed to generate refresh token", err)
	}

	session, err := entities.NewRefreshSession(
		uc.idGen.NewID(),
		cred.UserID,
		refreshHash,
		now.Add(uc.ttl.RefreshTTL),
		input.UserAgent,
		input.IPAddress,
		now,
	)
	if err != nil {
		return LoginUserOutput{}, apperrors.Internal("failed to build session", err)
	}

	if err := uc.sessions.Create(ctx, session); err != nil {
		return LoginUserOutput{}, apperrors.Internal("failed to persist session", err)
	}

	// Best effort: o login continua bem-sucedido mesmo com falha de publicação.
	_ = uc.events.PublishLoginSucceeded(ctx, ports.LoginSucceededEvent{
		UserID:         cred.UserID,
		LoggedAt:       now.UTC().Format(time.RFC3339Nano),
		CorrelationID:  input.CorrelationID,
		CausationID:    input.CausationID,
		Traceparent:    input.Traceparent,
		IdempotencyKey: input.IdempotencyKey,
	})

	return LoginUserOutput{
		UserID: cred.UserID,
		Tokens: valueobjects.AuthTokens{
			AccessToken:  accessToken,
			RefreshToken: plainRefresh,
			TokenType:    "Bearer",
			ExpiresIn:    uc.ttl.AccessSeconds(),
		},
	}, nil
}
