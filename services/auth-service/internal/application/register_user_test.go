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

func TestRegisterUserUseCase_Execute(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	validEmail := "user@example.com"
	validPassword := "Secret1@"

	buildExistingCred := func() *entities.Credential {
		email, _ := valueobjects.NewEmail(validEmail)
		c, _ := entities.NewCredential("cid", "uid", email, "hash", now)
		return c
	}

	tests := []struct {
		name     string
		input    application.RegisterUserInput
		credRepo *fakeCredentialRepo
		hasher   *fakeHasher
		wantCode string
		wantErr  bool
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
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uc := application.NewRegisterUserUseCase(
				tc.credRepo,
				tc.hasher,
				fixedClock{t: now},
				&seqIDGen{ids: []string{"cred-id-1", "user-id-1"}},
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
		})
	}
}
