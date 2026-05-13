package ports

import (
	"context"
	"time"

	"food_delivery_platform/services/auth-service/internal/domain/entities"
	"food_delivery_platform/services/auth-service/internal/domain/valueobjects"
)

type credentialRepoStub struct{}

func (credentialRepoStub) Create(ctx context.Context, credential *entities.Credential) error {
	return nil
}
func (credentialRepoStub) GetByEmail(ctx context.Context, email string) (*entities.Credential, error) {
	return nil, nil
}
func (credentialRepoStub) GetByUserID(ctx context.Context, userID string) (*entities.Credential, error) {
	return nil, nil
}
func (credentialRepoStub) Update(ctx context.Context, credential *entities.Credential) error {
	return nil
}

type refreshRepoStub struct{}

func (refreshRepoStub) Create(ctx context.Context, session *entities.RefreshSession) error {
	return nil
}
func (refreshRepoStub) GetByTokenHash(ctx context.Context, tokenHash string) (*entities.RefreshSession, error) {
	return nil, nil
}
func (refreshRepoStub) GetByID(ctx context.Context, id string) (*entities.RefreshSession, error) {
	return nil, nil
}
func (refreshRepoStub) Revoke(ctx context.Context, sessionID string) error {
	return nil
}
func (refreshRepoStub) RevokeAllByUserID(ctx context.Context, userID string) error {
	return nil
}
func (refreshRepoStub) Update(ctx context.Context, session *entities.RefreshSession) error {
	return nil
}

type hasherStub struct{}

func (hasherStub) Hash(ctx context.Context, plain string) (string, error) {
	return "", nil
}
func (hasherStub) Compare(ctx context.Context, plain, hash string) error {
	return nil
}

type tokenServiceStub struct{}

func (tokenServiceStub) GenerateAccessToken(ctx context.Context, claims valueobjects.TokenClaims) (string, error) {
	return "", nil
}
func (tokenServiceStub) ValidateAccessToken(ctx context.Context, token string) (valueobjects.TokenClaims, error) {
	return valueobjects.TokenClaims{}, nil
}
func (tokenServiceStub) GenerateRefreshToken(ctx context.Context) (string, string, error) {
	return "", "", nil
}

type clockStub struct{}

func (clockStub) Now() time.Time {
	return time.Now()
}

type idGenStub struct{}

func (idGenStub) NewID() string {
	return "id"
}

var _ CredentialRepository = (*credentialRepoStub)(nil)
var _ RefreshTokenRepository = (*refreshRepoStub)(nil)
var _ PasswordHasher = (*hasherStub)(nil)
var _ TokenService = (*tokenServiceStub)(nil)
var _ Clock = (*clockStub)(nil)
var _ IDGenerator = (*idGenStub)(nil)
