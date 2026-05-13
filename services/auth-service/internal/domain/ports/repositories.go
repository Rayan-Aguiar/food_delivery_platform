package ports

import (
	"context"

	"food_delivery_platform/services/auth-service/internal/domain/entities"
)

type CredentialRepository interface {
	Create(ctx context.Context, credential *entities.Credential) error
	GetByEmail(ctx context.Context, email string) (*entities.Credential, error)
	GetByUserID(ctx context.Context, userID string) (*entities.Credential, error)
	Update(ctx context.Context, credential *entities.Credential) error
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, session *entities.RefreshSession) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*entities.RefreshSession, error)
	GetByID(ctx context.Context, id string) (*entities.RefreshSession, error)
	Revoke(ctx context.Context, sessionID string) error
	RevokeAllByUserID(ctx context.Context, userID string) error
	Update(ctx context.Context, session *entities.RefreshSession) error
}
