package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	apperrors "food_delivery_platform/shared/errors"

	"food_delivery_platform/services/auth-service/internal/application"
	"food_delivery_platform/services/auth-service/internal/domain/entities"
	"food_delivery_platform/services/auth-service/internal/domain/ports"
	"food_delivery_platform/services/auth-service/internal/domain/valueobjects"
)

func mustTokenTTL(access, refresh time.Duration) valueobjects.TokenTTL {
	ttl, err := valueobjects.NewTokenTTL(access, refresh)
	if err != nil {
		panic(err)
	}
	return ttl
}

func mustActiveCred(t *testing.T, now time.Time) *entities.Credential {
	t.Helper()
	email, _ := valueobjects.NewEmail("user@example.com")
	c, err := entities.NewCredential("cred-1", "user-1", email, "hashed_Secret1@", now)
	if err != nil {
		t.Fatalf("build cred: %v", err)
	}
	return c
}

func TestLoginUserUseCase_Execute(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	ttl := mustTokenTTL(15*time.Minute, 7*24*time.Hour)

	validInput := application.LoginUserInput{
		Email:     "user@example.com",
		Password:  "Secret1@",
		UserAgent: "test-agent",
		IPAddress: "127.0.0.1",
	}

	tests := []struct {
		name         string
		input        application.LoginUserInput
		credRepo     *fakeCredentialRepo
		sessionRepo  *fakeSessionRepo
		hasher       *fakeHasher
		tokenService *fakeTokenService
		wantCode     string
		wantErr      bool
	}{
		{
			name:  "success",
			input: validInput,
			credRepo: &fakeCredentialRepo{
				getByEmailFn: func(_ context.Context, _ string) (*entities.Credential, error) {
					return mustActiveCred(t, now), nil
				},
			},
			sessionRepo:  &fakeSessionRepo{},
			hasher:       &fakeHasher{},
			tokenService: &fakeTokenService{},
		},
		{
			name:         "invalid email format",
			input:        application.LoginUserInput{Email: "not-an-email", Password: "Secret1@"},
			credRepo:     &fakeCredentialRepo{},
			sessionRepo:  &fakeSessionRepo{},
			hasher:       &fakeHasher{},
			tokenService: &fakeTokenService{},
			wantErr:      true,
			wantCode:     apperrors.CodeUnauthorized,
		},
		{
			name:  "credential not found",
			input: validInput,
			credRepo: &fakeCredentialRepo{
				getByEmailFn: func(_ context.Context, _ string) (*entities.Credential, error) {
					return nil, nil
				},
			},
			sessionRepo:  &fakeSessionRepo{},
			hasher:       &fakeHasher{},
			tokenService: &fakeTokenService{},
			wantErr:      true,
			wantCode:     apperrors.CodeUnauthorized,
		},
		{
			name:  "credential disabled",
			input: validInput,
			credRepo: &fakeCredentialRepo{
				getByEmailFn: func(_ context.Context, _ string) (*entities.Credential, error) {
					c := mustActiveCred(t, now)
					c.Disable(now)
					return c, nil
				},
			},
			sessionRepo:  &fakeSessionRepo{},
			hasher:       &fakeHasher{},
			tokenService: &fakeTokenService{},
			wantErr:      true,
			wantCode:     apperrors.CodeUnauthorized,
		},
		{
			name:  "wrong password records failed attempt",
			input: validInput,
			credRepo: &fakeCredentialRepo{
				getByEmailFn: func(_ context.Context, _ string) (*entities.Credential, error) {
					return mustActiveCred(t, now), nil
				},
				updateFn: func(_ context.Context, c *entities.Credential) error {
					if c.FailedLoginAttempts != 1 {
						t.Errorf("expected 1 failed attempt, got %d", c.FailedLoginAttempts)
					}
					return nil
				},
			},
			sessionRepo: &fakeSessionRepo{},
			hasher: &fakeHasher{
				compareFn: func(_ context.Context, _, _ string) error {
					return errors.New("wrong password")
				},
			},
			tokenService: &fakeTokenService{},
			wantErr:      true,
			wantCode:     apperrors.CodeUnauthorized,
		},
		{
			name:  "repo lookup error",
			input: validInput,
			credRepo: &fakeCredentialRepo{
				getByEmailFn: func(_ context.Context, _ string) (*entities.Credential, error) {
					return nil, errors.New("db timeout")
				},
			},
			sessionRepo:  &fakeSessionRepo{},
			hasher:       &fakeHasher{},
			tokenService: &fakeTokenService{},
			wantErr:      true,
			wantCode:     apperrors.CodeInternal,
		},
		{
			name:  "access token generation fails",
			input: validInput,
			credRepo: &fakeCredentialRepo{
				getByEmailFn: func(_ context.Context, _ string) (*entities.Credential, error) {
					return mustActiveCred(t, now), nil
				},
			},
			sessionRepo: &fakeSessionRepo{},
			hasher:      &fakeHasher{},
			tokenService: &fakeTokenService{
				generateAccessFn: func(_ context.Context, _ valueobjects.TokenClaims) (string, error) {
					return "", errors.New("signing key missing")
				},
			},
			wantErr:  true,
			wantCode: apperrors.CodeInternal,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uc := application.NewLoginUserUseCase(
				tc.credRepo,
				tc.sessionRepo,
				tc.hasher,
				tc.tokenService,
				fixedClock{t: now},
				&seqIDGen{ids: []string{"session-id-1"}},
				ttl,
			)

			out, err := uc.Execute(ctx, tc.input)

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				var appErr *apperrors.AppError
				if !errors.As(err, &appErr) {
					t.Fatalf("expected *AppError, got %T: %v", err, err)
				}
				if appErr.Code != tc.wantCode {
					t.Errorf("expected code %q, got %q", tc.wantCode, appErr.Code)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if out.UserID == "" {
				t.Error("expected non-empty UserID")
			}
			if out.Tokens.AccessToken == "" {
				t.Error("expected non-empty AccessToken")
			}
			if out.Tokens.RefreshToken == "" {
				t.Error("expected non-empty RefreshToken")
			}
			if out.Tokens.TokenType != "Bearer" {
				t.Errorf("expected TokenType Bearer, got %q", out.Tokens.TokenType)
			}
		})
	}
}

func TestLoginUserUseCase_PublishesLoginSucceededEvent(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	ttl := mustTokenTTL(15*time.Minute, 7*24*time.Hour)
	eventsPublisher := &fakeAuthEventPublisher{}

	uc := application.NewLoginUserUseCase(
		&fakeCredentialRepo{getByEmailFn: func(_ context.Context, _ string) (*entities.Credential, error) {
			return mustActiveCred(t, now), nil
		}},
		&fakeSessionRepo{},
		&fakeHasher{},
		&fakeTokenService{},
		fixedClock{t: now},
		&seqIDGen{ids: []string{"session-id-1"}},
		ttl,
		eventsPublisher,
	)

	_, err := uc.Execute(ctx, application.LoginUserInput{
		Email:          "user@example.com",
		Password:       "Secret1@",
		CorrelationID:  "corr-1",
		CausationID:    "req-1",
		Traceparent:    "00-aabbccddeeff00112233445566778899-0011223344556677-01",
		IdempotencyKey: "idem-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(eventsPublisher.loginCalls) != 1 {
		t.Fatalf("expected 1 published login event, got %d", len(eventsPublisher.loginCalls))
	}
	got := eventsPublisher.loginCalls[0]
	if got.UserID != "user-1" {
		t.Errorf("user id = %q, want %q", got.UserID, "user-1")
	}
	if got.CorrelationID != "corr-1" {
		t.Errorf("correlation id = %q, want %q", got.CorrelationID, "corr-1")
	}
}

func TestLoginUserUseCase_PublishFailureIsBestEffort(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	ttl := mustTokenTTL(15*time.Minute, 7*24*time.Hour)
	eventsPublisher := &fakeAuthEventPublisher{
		publishLoginFn: func(context.Context, ports.LoginSucceededEvent) error { return failPublisherError() },
	}

	uc := application.NewLoginUserUseCase(
		&fakeCredentialRepo{getByEmailFn: func(_ context.Context, _ string) (*entities.Credential, error) {
			return mustActiveCred(t, now), nil
		}},
		&fakeSessionRepo{},
		&fakeHasher{},
		&fakeTokenService{},
		fixedClock{t: now},
		&seqIDGen{ids: []string{"session-id-1"}},
		ttl,
		eventsPublisher,
	)

	out, err := uc.Execute(ctx, application.LoginUserInput{Email: "user@example.com", Password: "Secret1@"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.UserID == "" {
		t.Fatal("expected successful login output")
	}
	if len(eventsPublisher.loginCalls) != 1 {
		t.Fatalf("expected 1 publish attempt, got %d", len(eventsPublisher.loginCalls))
	}
}
