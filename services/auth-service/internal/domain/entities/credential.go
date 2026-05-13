package entities

import (
	"time"

	domainerrors "food_delivery_platform/services/auth-service/internal/domain/errors"
	"food_delivery_platform/services/auth-service/internal/domain/valueobjects"
)

type Credential struct {
	ID                  string
	UserID              string
	Email               valueobjects.Email
	PasswordHash        string
	Status              valueobjects.CredentialStatus
	FailedLoginAttempts int
	LastLoginAt         *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

func NewCredential(
	id string,
	userID string,
	email valueobjects.Email,
	passwordHash string,
	now time.Time,
) (*Credential, error) {
	if id == "" {
		return nil, domainerrors.ErrEmptyID
	}
	if userID == "" {
		return nil, domainerrors.ErrEmptyUserID
	}
	if passwordHash == "" {
		return nil, domainerrors.ErrEmptyPasswordHash
	}

	return &Credential{
		ID:                  id,
		UserID:              userID,
		Email:               email,
		PasswordHash:        passwordHash,
		Status:              valueobjects.CredentialStatusActive,
		FailedLoginAttempts: 0,
		CreatedAt:           now,
		UpdatedAt:           now,
	}, nil
}

func (c *Credential) CanLogin() bool {
	return c.Status == valueobjects.CredentialStatusActive
}

func (c *Credential) RegisterFailedAttempt(now time.Time) {
	c.FailedLoginAttempts++
	c.UpdatedAt = now
}

func (c *Credential) RegisterSuccessfulLogin(now time.Time) error {
	if !c.CanLogin() {
		return domainerrors.ErrCredentialDisabled
	}
	c.FailedLoginAttempts = 0
	c.LastLoginAt = &now
	c.UpdatedAt = now
	return nil
}

func (c *Credential) ChangePassword(newHash string, now time.Time) error {
	if newHash == "" {
		return domainerrors.ErrEmptyPasswordHash
	}
	c.PasswordHash = newHash
	c.UpdatedAt = now
	return nil
}

func (c *Credential) Disable(now time.Time) {
	c.Status = valueobjects.CredentialStatusDisabled
	c.UpdatedAt = now
}
