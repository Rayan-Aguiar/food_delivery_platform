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

func TestRegisterUserUseCase_Execute(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	validEmail := "user@example.com"
	validPassword := "Secret1@"

	ttl, _ := valueobjects.NewTokenTTL(15*time.Minute, 7*24*time.Hour)

	buildExistingCred := func() *entities.Credential {
		email, _ := valueobjects.NewEmail(validEmail)
		c, _ := entities.NewCredential("cid", "uid", email, "hash", now)
		return c
	}

	tests := []struct {
		name        string
		input       application.RegisterUserInput
		credRepo    *fakeCredentialRepo
		sessionRepo *fakeSessionRepo
		hasher      *fakeHasher
		tokenSvc    *fakeTokenService
		wantCode    string
		wantErr     bool
	}{
		{
			name:     "success",
			input:    application.RegisterUserInput{Email: validEmail, Password: validPassword},
			credRepo: &fakeCredentialRepo{},
			hasher:   &fakeHasher{},
		},
		{
			name:     "invalid email",
			input:    application.RegisterUserInput{Email: "not-an-email", Password: validPassword},
			credRepo: &fakeCredentialRepo{},
			hasher:   &fakeHasher{},
			wantErr:  true,
			wantCode: apperrors.CodeInvalidArgument,
		},
		{
			name:     "weak password",
			input:    application.RegisterUserInput{Email: validEmail, Password: "weak"},
			credRepo: &fakeCredentialRepo{},
			hasher:   &fakeHasher{},
			wantErr:  true,
			wantCode: apperrors.CodeInvalidArgument,
		},
		{
			name:  "email already registered",
			input: application.RegisterUserInput{Email: validEmail, Password: validPassword},
			credRepo: &fakeCredentialRepo{
				getByEmailFn: func(_ context.Context, _ string) (*entities.Credential, error) {
					return buildExistingCred(), nil
				},
			},
			hasher:   &fakeHasher{},
			wantErr:  true,
			wantCode: apperrors.CodeConflict,
		},
		{
			name:  "repo lookup error",
			input: application.RegisterUserInput{Email: validEmail, Password: validPassword},
			credRepo: &fakeCredentialRepo{
				getByEmailFn: func(_ context.Context, _ string) (*entities.Credential, error) {
					return nil, errors.New("db down")
				},
			},
			hasher:   &fakeHasher{},
			wantErr:  true,
			wantCode: apperrors.CodeInternal,
		},
		{
			name:     "hash error",
			input:    application.RegisterUserInput{Email: validEmail, Password: validPassword},
			credRepo: &fakeCredentialRepo{},
			hasher: &fakeHasher{
				hashFn: func(_ context.Context, _ string) (string, error) {
					return "", errors.New("bcrypt failed")
				},
			},
			wantErr:  true,
			wantCode: apperrors.CodeInternal,
		},
		{
			name:  "repo create error",
			input: application.RegisterUserInput{Email: validEmail, Password: validPassword},
			credRepo: &fakeCredentialRepo{
				createFn: func(_ context.Context, _ *entities.Credential) error {
					return errors.New("insert failed")
				},
			},
			hasher:   &fakeHasher{},
			wantErr:  true,
			wantCode: apperrors.CodeInternal,
		},
		{
			name:     "access token generation error",
			input:    application.RegisterUserInput{Email: validEmail, Password: validPassword},
			credRepo: &fakeCredentialRepo{},
			hasher:   &fakeHasher{},
			tokenSvc: &fakeTokenService{
				generateAccessFn: func(_ context.Context, _ valueobjects.TokenClaims) (string, error) {
					return "", errors.New("jwt failed")
				},
			},
			wantErr:  true,
			wantCode: apperrors.CodeInternal,
		},
		{
			name:     "refresh token generation error",
			input:    application.RegisterUserInput{Email: validEmail, Password: validPassword},
			credRepo: &fakeCredentialRepo{},
			hasher:   &fakeHasher{},
			tokenSvc: &fakeTokenService{
				generateRefreshFn: func(_ context.Context) (string, string, error) {
					return "", "", errors.New("rand failed")
				},
			},
			wantErr:  true,
			wantCode: apperrors.CodeInternal,
		},
		{
			name:     "session create error",
			input:    application.RegisterUserInput{Email: validEmail, Password: validPassword},
			credRepo: &fakeCredentialRepo{},
			hasher:   &fakeHasher{},
			sessionRepo: &fakeSessionRepo{
				createFn: func(_ context.Context, _ *entities.RefreshSession) error {
					return errors.New("session insert failed")
				},
			},
			wantErr:  true,
			wantCode: apperrors.CodeInternal,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sessionRepo := tc.sessionRepo
			if sessionRepo == nil {
				sessionRepo = &fakeSessionRepo{}
			}
			tokenSvc := tc.tokenSvc
			if tokenSvc == nil {
				tokenSvc = &fakeTokenService{}
			}

			uc := application.NewRegisterUserUseCase(
				tc.credRepo,
				sessionRepo,
				tc.hasher,
				tokenSvc,
				fixedClock{t: now},
				&seqIDGen{ids: []string{"cred-id-1", "user-id-1", "session-id-1"}},
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
			if out.CredentialID == "" {
				t.Error("expected non-empty CredentialID")
			}
			if out.Tokens.AccessToken == "" {
				t.Error("expected non-empty access token")
			}
			if out.Tokens.RefreshToken == "" {
				t.Error("expected non-empty refresh token")
			}
		})
	}
}

func TestRegisterUserUseCase_PublishesUserRegisteredEvent(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	ttl, _ := valueobjects.NewTokenTTL(15*time.Minute, 7*24*time.Hour)
	eventsPublisher := &fakeAuthEventPublisher{}

	uc := application.NewRegisterUserUseCase(
		&fakeCredentialRepo{},
		&fakeSessionRepo{},
		&fakeHasher{},
		&fakeTokenService{},
		fixedClock{t: now},
		&seqIDGen{ids: []string{"cred-id-1", "user-id-1", "session-id-1"}},
		ttl,
		eventsPublisher,
	)

	_, err := uc.Execute(ctx, application.RegisterUserInput{
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

	if len(eventsPublisher.registeredCalls) != 1 {
		t.Fatalf("expected 1 published register event, got %d", len(eventsPublisher.registeredCalls))
	}
	got := eventsPublisher.registeredCalls[0]
	if got.UserID != "user-id-1" {
		t.Errorf("user id = %q, want %q", got.UserID, "user-id-1")
	}
	if got.Email != "user@example.com" {
		t.Errorf("email = %q, want %q", got.Email, "user@example.com")
	}
	if got.CorrelationID != "corr-1" {
		t.Errorf("correlation id = %q, want %q", got.CorrelationID, "corr-1")
	}
}

func TestRegisterUserUseCase_PublishFailureIsBestEffort(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	ttl, _ := valueobjects.NewTokenTTL(15*time.Minute, 7*24*time.Hour)
	eventsPublisher := &fakeAuthEventPublisher{
		publishRegisteredFn: func(context.Context, ports.UserRegisteredEvent) error { return failPublisherError() },
	}

	uc := application.NewRegisterUserUseCase(
		&fakeCredentialRepo{},
		&fakeSessionRepo{},
		&fakeHasher{},
		&fakeTokenService{},
		fixedClock{t: now},
		&seqIDGen{ids: []string{"cred-id-1", "user-id-1", "session-id-1"}},
		ttl,
		eventsPublisher,
	)

	out, err := uc.Execute(ctx, application.RegisterUserInput{Email: "user@example.com", Password: "Secret1@"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.UserID == "" {
		t.Fatal("expected successful register output")
	}
	if len(eventsPublisher.registeredCalls) != 1 {
		t.Fatalf("expected 1 publish attempt, got %d", len(eventsPublisher.registeredCalls))
	}
}
