package valueobjects

import (
	"time"

	domainerrors "food_delivery_platform/services/auth-service/internal/domain/errors"
)

type TokenTTL struct {
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

func NewTokenTTL(accessTTL, refreshTTL time.Duration) (TokenTTL, error) {
	if accessTTL <= 0 || refreshTTL <= 0 || refreshTTL <= accessTTL {
		return TokenTTL{}, domainerrors.ErrInvalidTokenTTL
	}
	return TokenTTL{
		AccessTTL:  accessTTL,
		RefreshTTL: refreshTTL,
	}, nil
}

func (t TokenTTL) AccessSeconds() int64 {
	return int64(t.AccessTTL.Seconds())
}
