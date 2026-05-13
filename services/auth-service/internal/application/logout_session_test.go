package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	apperrors "food_delivery_platform/shared/errors"

	"food_delivery_platform/services/auth-service/internal/application"
	"food_delivery_platform/services/auth-service/internal/domain/entities"
)

func TestLogoutSessionUseCase_Execute(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 1, 3, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		input        application.LogoutSessionInput
		sessionRepo  *fakeSessionRepo
		tokenService *fakeTokenService
		wantErr      bool
		wantCode     string
	}{
		{
			name:  "success",
			input: application.LogoutSessionInput{RefreshToken: "plain_refresh"},
			sessionRepo: &fakeSessionRepo{
				getByTokenHashFn: func(_ context.Context, _ string) (*entities.RefreshSession, error) { return mustSession(t, now), nil },
			},
			tokenService: &fakeTokenService{
				hashRefreshFn: func(_ context.Context, _ string) (string, error) { return "hashed_refresh_token", nil },
			},
		},
		{
			name:         "empty token",
			input:        application.LogoutSessionInput{RefreshToken: ""},
			sessionRepo:  &fakeSessionRepo{},
			tokenService: &fakeTokenService{},
			wantErr:      true,
			wantCode:     apperrors.CodeUnauthorized,
		},
		{
			name:        "token hash fails",
			input:       application.LogoutSessionInput{RefreshToken: "plain_refresh"},
			sessionRepo: &fakeSessionRepo{},
			tokenService: &fakeTokenService{
				hashRefreshFn: func(_ context.Context, _ string) (string, error) { return "", errors.New("hash failed") },
			},
			wantErr:  true,
			wantCode: apperrors.CodeInternal,
		},
		{
			name:  "session not found",
			input: application.LogoutSessionInput{RefreshToken: "plain_refresh"},
			sessionRepo: &fakeSessionRepo{
				getByTokenHashFn: func(_ context.Context, _ string) (*entities.RefreshSession, error) { return nil, nil },
			},
			tokenService: &fakeTokenService{
				hashRefreshFn: func(_ context.Context, _ string) (string, error) { return "hashed_refresh_token", nil },
			},
			wantErr:  true,
			wantCode: apperrors.CodeUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uc := application.NewLogoutSessionUseCase(tc.sessionRepo, tc.tokenService)

			err := uc.Execute(ctx, tc.input)

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
		})
	}
}
