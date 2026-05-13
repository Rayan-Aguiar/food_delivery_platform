package entities

import (
	"time"

	domainerrors "food_delivery_platform/services/auth-service/internal/domain/errors"
)

type RefreshSession struct {
	ID                   string
	UserID               string
	TokenHash            string
	ExpiresAt            time.Time
	RevokedAt            *time.Time
	RotatedFromSessionID *string
	UserAgent            string
	IPAddress            string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

func NewRefreshSession(
	id string,
	userID string,
	tokenHash string,
	expiresAt time.Time,
	userAgent string,
	ipAddress string,
	now time.Time,
) (*RefreshSession, error) {
	if id == "" {
		return nil, domainerrors.ErrEmptyID
	}
	if userID == "" {
		return nil, domainerrors.ErrEmptyUserID
	}
	if tokenHash == "" {
		return nil, domainerrors.ErrEmptyTokenHash
	}
	if expiresAt.IsZero() || !expiresAt.After(now) {
		return nil, domainerrors.ErrInvalidTokenTTL
	}

	return &RefreshSession{
		ID:        id,
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
		UserAgent: userAgent,
		IPAddress: ipAddress,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (s *RefreshSession) IsExpired(now time.Time) bool {
	return now.After(s.ExpiresAt)
}

func (s *RefreshSession) IsRevoked() bool {
	return s.RevokedAt != nil
}

func (s *RefreshSession) CanBeUsed(now time.Time) error {
	if s.IsRevoked() {
		return domainerrors.ErrRefreshTokenRevoked
	}
	if s.IsExpired(now) {
		return domainerrors.ErrRefreshTokenExpired
	}
	return nil
}

func (s *RefreshSession) Revoke(now time.Time) {
	s.RevokedAt = &now
	s.UpdatedAt = now
}

func (s *RefreshSession) Rotate(
	newID string,
	newTokenHash string,
	newExpiresAt time.Time,
	now time.Time,
) (*RefreshSession, error) {
	if err := s.CanBeUsed(now); err != nil {
		return nil, err
	}
	if newID == "" {
		return nil, domainerrors.ErrEmptyID
	}
	if newTokenHash == "" {
		return nil, domainerrors.ErrEmptyTokenHash
	}
	if !newExpiresAt.After(now) {
		return nil, domainerrors.ErrInvalidTokenTTL
	}

	s.Revoke(now)
	rotatedFrom := s.ID

	return &RefreshSession{
		ID:                   newID,
		UserID:               s.UserID,
		TokenHash:            newTokenHash,
		ExpiresAt:            newExpiresAt,
		RotatedFromSessionID: &rotatedFrom,
		UserAgent:            s.UserAgent,
		IPAddress:            s.IPAddress,
		CreatedAt:            now,
		UpdatedAt:            now,
	}, nil
}
