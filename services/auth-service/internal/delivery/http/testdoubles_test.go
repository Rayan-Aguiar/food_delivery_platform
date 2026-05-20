package httpdelivery

import (
	"context"

	"food_delivery_platform/services/auth-service/internal/application"
	"food_delivery_platform/services/auth-service/internal/domain/valueobjects"
)

// defaultTokens é retornado pelos fakes quando não há fn customizada.
var defaultTokens = valueobjects.AuthTokens{
	AccessToken:  "access.token.here",
	RefreshToken: "refresh-token-value",
	TokenType:    "Bearer",
	ExpiresIn:    900,
}

// ── register ─────────────────────────────────────────────────────────────────

type fakeRegister struct {
	fn func(context.Context, application.RegisterUserInput) (application.RegisterUserOutput, error)
}

func (f *fakeRegister) Execute(ctx context.Context, in application.RegisterUserInput) (application.RegisterUserOutput, error) {
	if f.fn != nil {
		return f.fn(ctx, in)
	}
	return application.RegisterUserOutput{UserID: "user-1", CredentialID: "cred-1", Tokens: defaultTokens}, nil
}

// ── login ─────────────────────────────────────────────────────────────────────

type fakeLogin struct {
	fn func(context.Context, application.LoginUserInput) (application.LoginUserOutput, error)
}

func (f *fakeLogin) Execute(ctx context.Context, in application.LoginUserInput) (application.LoginUserOutput, error) {
	if f.fn != nil {
		return f.fn(ctx, in)
	}
	return application.LoginUserOutput{UserID: "user-1", Tokens: defaultTokens}, nil
}

// ── refresh ───────────────────────────────────────────────────────────────────

type fakeRefresh struct {
	fn func(context.Context, application.RefreshAccessTokenInput) (application.RefreshAccessTokenOutput, error)
}

func (f *fakeRefresh) Execute(ctx context.Context, in application.RefreshAccessTokenInput) (application.RefreshAccessTokenOutput, error) {
	if f.fn != nil {
		return f.fn(ctx, in)
	}
	return application.RefreshAccessTokenOutput{UserID: "user-1", Tokens: defaultTokens}, nil
}

// ── logout ────────────────────────────────────────────────────────────────────

type fakeLogout struct {
	fn func(context.Context, application.LogoutSessionInput) error
}

func (f *fakeLogout) Execute(ctx context.Context, in application.LogoutSessionInput) error {
	if f.fn != nil {
		return f.fn(ctx, in)
	}
	return nil
}
