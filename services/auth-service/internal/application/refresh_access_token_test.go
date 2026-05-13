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

func mustSession(t *testing.T, now time.Time) *entities.RefreshSession {
	t.Helper()
	s, err := entities.NewRefreshSession(
		"session-1",
		"user-1",
		"hashed_refresh_token",
		now.Add(24*time.Hour),
		"ua",
		"127.0.0.1",
		now,
	)
	if err != nil {
		t.Fatalf("build session: %v", err)
	}
	return s
}

func TestRefreshAccessTokenUseCase_Execute(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 1, 2, 12, 0, 0, 0, time.UTC)
	ttl := mustTokenTTL(15*time.Minute, 7*24*time.Hour)

	activeCred := mustActiveCred(t, now)

	tests := []struct {
		name         string
		input        application.RefreshAccessTokenInput
		credRepo     *fakeCredentialRepo
		sessionRepo  *fakeSessionRepo
		tokenService *fakeTokenService
		wantErr      bool
		wantCode     string
	}{
		{
			name:  "success",
			input: application.RefreshAccessTokenInput{RefreshToken: "plain_refresh"},
			credRepo: &fakeCredentialRepo{
				getByUserIDFn: func(_ context.Context, _ string) (*entities.Credential, error) { return activeCred, nil },
			},
			sessionRepo: &fakeSessionRepo{
				getByTokenHashFn: func(_ context.Context, _ string) (*entities.RefreshSession, error) { return mustSession(t, now), nil },
			},
			tokenService: &fakeTokenService{
				hashRefreshFn: func(_ context.Context, _ string) (string, error) { return "hashed_refresh_token", nil },
				generateRefreshFn: func(_ context.Context) (string, string, error) {
					return "new_plain_refresh", "new_hashed_refresh", nil
				},
			},
		},
		{
			name:         "empty token",
			input:        application.RefreshAccessTokenInput{RefreshToken: ""},
			credRepo:     &fakeCredentialRepo{},
			sessionRepo:  &fakeSessionRepo{},
			tokenService: &fakeTokenService{},
			wantErr:      true,
			wantCode:     apperrors.CodeUnauthorized,
		},
		{
			name:        "hash error",
			input:       application.RefreshAccessTokenInput{RefreshToken: "plain_refresh"},
			credRepo:    &fakeCredentialRepo{},
			sessionRepo: &fakeSessionRepo{},
			tokenService: &fakeTokenService{
				hashRefreshFn: func(_ context.Context, _ string) (string, error) { return "", errors.New("hash failed") },
			},
			wantErr:  true,
			wantCode: apperrors.CodeInternal,
		},
		{
			name:     "session not found",
			input:    application.RefreshAccessTokenInput{RefreshToken: "plain_refresh"},
			credRepo: &fakeCredentialRepo{},
			sessionRepo: &fakeSessionRepo{
				getByTokenHashFn: func(_ context.Context, _ string) (*entities.RefreshSession, error) { return nil, nil },
			},
			tokenService: &fakeTokenService{
				hashRefreshFn: func(_ context.Context, _ string) (string, error) { return "hashed_refresh_token", nil },
			},
			wantErr:  true,
			wantCode: apperrors.CodeUnauthorized,
		},
		{
			name:     "session expired",
			input:    application.RefreshAccessTokenInput{RefreshToken: "plain_refresh"},
			credRepo: &fakeCredentialRepo{},
			sessionRepo: &fakeSessionRepo{
				getByTokenHashFn: func(_ context.Context, _ string) (*entities.RefreshSession, error) {
					s, _ := entities.NewRefreshSession("session-1", "user-1", "hashed_refresh_token", now.Add(-time.Minute), "ua", "127.0.0.1", now.Add(-2*time.Minute))
					return s, nil
				},
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
			uc := application.NewRefreshAccessTokenUseCase(
				tc.credRepo,
				tc.sessionRepo,
				tc.tokenService,
				fixedClock{t: now},
				&seqIDGen{ids: []string{"session-2"}},
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
			if out.Tokens.AccessToken == "" || out.Tokens.RefreshToken == "" {
				t.Error("expected non-empty tokens")
			}
		})
	}
}
