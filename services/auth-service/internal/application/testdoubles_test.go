package application_test

import (
	"context"
	"errors"
	"time"

	"food_delivery_platform/services/auth-service/internal/domain/entities"
	"food_delivery_platform/services/auth-service/internal/domain/ports"
	"food_delivery_platform/services/auth-service/internal/domain/valueobjects"
)

// ── Credential Repository ────────────────────────────────────────────────────

type fakeCredentialRepo struct {
	createFn      func(ctx context.Context, c *entities.Credential) error
	getByEmailFn  func(ctx context.Context, email string) (*entities.Credential, error)
	getByUserIDFn func(ctx context.Context, userID string) (*entities.Credential, error)
	updateFn      func(ctx context.Context, c *entities.Credential) error
}

func (f *fakeCredentialRepo) Create(ctx context.Context, c *entities.Credential) error {
	if f.createFn != nil {
		return f.createFn(ctx, c)
	}
	return nil
}

func (f *fakeCredentialRepo) GetByEmail(ctx context.Context, email string) (*entities.Credential, error) {
	if f.getByEmailFn != nil {
		return f.getByEmailFn(ctx, email)
	}
	return nil, nil
}

func (f *fakeCredentialRepo) GetByUserID(ctx context.Context, userID string) (*entities.Credential, error) {
	if f.getByUserIDFn != nil {
		return f.getByUserIDFn(ctx, userID)
	}
	return nil, nil
}

func (f *fakeCredentialRepo) Update(ctx context.Context, c *entities.Credential) error {
	if f.updateFn != nil {
		return f.updateFn(ctx, c)
	}
	return nil
}

// ── Refresh Token Repository ─────────────────────────────────────────────────

type fakeSessionRepo struct {
	createFn          func(ctx context.Context, s *entities.RefreshSession) error
	getByTokenHashFn  func(ctx context.Context, hash string) (*entities.RefreshSession, error)
	getByIDFn         func(ctx context.Context, id string) (*entities.RefreshSession, error)
	revokeFn          func(ctx context.Context, id string) error
	revokeAllByUserFn func(ctx context.Context, userID string) error
	updateFn          func(ctx context.Context, s *entities.RefreshSession) error
}

func (f *fakeSessionRepo) Create(ctx context.Context, s *entities.RefreshSession) error {
	if f.createFn != nil {
		return f.createFn(ctx, s)
	}
	return nil
}

func (f *fakeSessionRepo) GetByTokenHash(ctx context.Context, hash string) (*entities.RefreshSession, error) {
	if f.getByTokenHashFn != nil {
		return f.getByTokenHashFn(ctx, hash)
	}
	return nil, nil
}

func (f *fakeSessionRepo) GetByID(ctx context.Context, id string) (*entities.RefreshSession, error) {
	if f.getByIDFn != nil {
		return f.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (f *fakeSessionRepo) Revoke(ctx context.Context, id string) error {
	if f.revokeFn != nil {
		return f.revokeFn(ctx, id)
	}
	return nil
}

func (f *fakeSessionRepo) RevokeAllByUserID(ctx context.Context, userID string) error {
	if f.revokeAllByUserFn != nil {
		return f.revokeAllByUserFn(ctx, userID)
	}
	return nil
}

func (f *fakeSessionRepo) Update(ctx context.Context, s *entities.RefreshSession) error {
	if f.updateFn != nil {
		return f.updateFn(ctx, s)
	}
	return nil
}

// ── Password Hasher ───────────────────────────────────────────────────────────

type fakeHasher struct {
	hashFn    func(ctx context.Context, plain string) (string, error)
	compareFn func(ctx context.Context, plain, hash string) error
}

func (f *fakeHasher) Hash(ctx context.Context, plain string) (string, error) {
	if f.hashFn != nil {
		return f.hashFn(ctx, plain)
	}
	return "hashed_" + plain, nil
}

func (f *fakeHasher) Compare(ctx context.Context, plain, hash string) error {
	if f.compareFn != nil {
		return f.compareFn(ctx, plain, hash)
	}
	return nil
}

// ── Token Service ─────────────────────────────────────────────────────────────

type fakeTokenService struct {
	generateAccessFn  func(ctx context.Context, claims valueobjects.TokenClaims) (string, error)
	validateAccessFn  func(ctx context.Context, token string) (valueobjects.TokenClaims, error)
	generateRefreshFn func(ctx context.Context) (string, string, error)
	hashRefreshFn     func(ctx context.Context, plain string) (string, error)
}

func (f *fakeTokenService) GenerateAccessToken(ctx context.Context, claims valueobjects.TokenClaims) (string, error) {
	if f.generateAccessFn != nil {
		return f.generateAccessFn(ctx, claims)
	}
	return "access_token", nil
}

func (f *fakeTokenService) ValidateAccessToken(ctx context.Context, token string) (valueobjects.TokenClaims, error) {
	if f.validateAccessFn != nil {
		return f.validateAccessFn(ctx, token)
	}
	return valueobjects.TokenClaims{}, nil
}

func (f *fakeTokenService) GenerateRefreshToken(ctx context.Context) (string, string, error) {
	if f.generateRefreshFn != nil {
		return f.generateRefreshFn(ctx)
	}
	return "plain_refresh", "hashed_refresh", nil
}

func (f *fakeTokenService) HashRefreshToken(ctx context.Context, plain string) (string, error) {
	if f.hashRefreshFn != nil {
		return f.hashRefreshFn(ctx, plain)
	}
	return "hashed_" + plain, nil
}

// ── Clock ─────────────────────────────────────────────────────────────────────

type fixedClock struct{ t time.Time }

func (c fixedClock) Now() time.Time { return c.t }

// ── ID Generator ──────────────────────────────────────────────────────────────

type seqIDGen struct {
	ids []string
	i   int
}

func (g *seqIDGen) NewID() string {
	id := g.ids[g.i%len(g.ids)]
	g.i++
	return id
}

// ── Auth Event Publisher ─────────────────────────────────────────────────────

type fakeAuthEventPublisher struct {
	publishRegisteredFn func(context.Context, ports.UserRegisteredEvent) error
	publishLoginFn      func(context.Context, ports.LoginSucceededEvent) error

	registeredCalls []ports.UserRegisteredEvent
	loginCalls      []ports.LoginSucceededEvent
}

func (f *fakeAuthEventPublisher) PublishUserRegistered(ctx context.Context, in ports.UserRegisteredEvent) error {
	f.registeredCalls = append(f.registeredCalls, in)
	if f.publishRegisteredFn != nil {
		return f.publishRegisteredFn(ctx, in)
	}
	return nil
}

func (f *fakeAuthEventPublisher) PublishLoginSucceeded(ctx context.Context, in ports.LoginSucceededEvent) error {
	f.loginCalls = append(f.loginCalls, in)
	if f.publishLoginFn != nil {
		return f.publishLoginFn(ctx, in)
	}
	return nil
}

func failPublisherError() error {
	return errors.New("publish failed")
}
