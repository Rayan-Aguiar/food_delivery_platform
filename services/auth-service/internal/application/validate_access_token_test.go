package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	apperrors "food_delivery_platform/shared/errors"

	"food_delivery_platform/services/auth-service/internal/application"
	"food_delivery_platform/services/auth-service/internal/domain/entities"
	"food_delivery_platform/services/auth-service/internal/domain/valueobjects"
)

func TestValidateAccessTokenUseCase_Execute(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 1, 4, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		input        application.ValidateAccessTokenInput
		credRepo     *fakeCredentialRepo
		tokenService *fakeTokenService
		wantErr      bool
		wantCode     string
	}{
		{
			name:  "success",
			input: application.ValidateAccessTokenInput{AccessToken: "valid_access"},
			credRepo: &fakeCredentialRepo{
				getByUserIDFn: func(_ context.Context, _ string) (*entities.Credential, error) { return mustActiveCred(t, now), nil },
			},
			tokenService: &fakeTokenService{
				validateAccessFn: func(_ context.Context, _ string) (valueobjects.TokenClaims, error) {
					return valueobjects.TokenClaims{Subject: "user-1", IssuedAt: now.Add(-time.Minute), ExpiresAt: now.Add(15 * time.Minute)}, nil
				},
			},
		},
		{
			name:         "empty token",
			input:        application.ValidateAccessTokenInput{AccessToken: ""},
			credRepo:     &fakeCredentialRepo{},
			tokenService: &fakeTokenService{},
			wantErr:      true,
			wantCode:     apperrors.CodeUnauthorized,
		},
		{
			name:     "invalid token signature",
			input:    application.ValidateAccessTokenInput{AccessToken: "bad_access"},
			credRepo: &fakeCredentialRepo{},
			tokenService: &fakeTokenService{
				validateAccessFn: func(_ context.Context, _ string) (valueobjects.TokenClaims, error) {
					return valueobjects.TokenClaims{}, errors.New("invalid signature")
				},
			},
			wantErr:  true,
			wantCode: apperrors.CodeUnauthorized,
		},
		{
			name:     "expired token",
			input:    application.ValidateAccessTokenInput{AccessToken: "expired_access"},
			credRepo: &fakeCredentialRepo{},
			tokenService: &fakeTokenService{
				validateAccessFn: func(_ context.Context, _ string) (valueobjects.TokenClaims, error) {
					return valueobjects.TokenClaims{Subject: "user-1", IssuedAt: now.Add(-2 * time.Hour), ExpiresAt: now.Add(-time.Minute)}, nil
				},
			},
			wantErr:  true,
			wantCode: apperrors.CodeUnauthorized,
		},
		{
			name:  "credential disabled",
			input: application.ValidateAccessTokenInput{AccessToken: "valid_access"},
			credRepo: &fakeCredentialRepo{
				getByUserIDFn: func(_ context.Context, _ string) (*entities.Credential, error) {
					c := mustActiveCred(t, now)
					c.Disable(now)
					return c, nil
				},
			},
			tokenService: &fakeTokenService{
				validateAccessFn: func(_ context.Context, _ string) (valueobjects.TokenClaims, error) {
					return valueobjects.TokenClaims{Subject: "user-1", IssuedAt: now.Add(-time.Minute), ExpiresAt: now.Add(15 * time.Minute)}, nil
				},
			},
			wantErr:  true,
			wantCode: apperrors.CodeUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uc := application.NewValidateAccessTokenUseCase(tc.credRepo, tc.tokenService, fixedClock{t: now})

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
			if out.Claims.Subject == "" {
				t.Error("expected non-empty claims subject")
			}
		})
	}
}
