package valueobjects

import (
	"unicode"

	domainerrors "food_delivery_platform/services/auth-service/internal/domain/errors"
)

type PasswordPolicy struct {
	MinLength      int
	RequireUpper   bool
	RequireLower   bool
	RequireNumber  bool
	RequireSpecial bool
}

func NewDefaultPasswordPolicy() PasswordPolicy {
	return PasswordPolicy{
		MinLength:      8,
		RequireUpper:   true,
		RequireLower:   true,
		RequireNumber:  true,
		RequireSpecial: true,
	}
}

func (p PasswordPolicy) Validate(password string) error {
	if p.MinLength <= 0 {
		return domainerrors.ErrInvalidPasswordPolicy
	}
	if len(password) < p.MinLength {
		return domainerrors.ErrWeakPassword
	}

	hasUpper := false
	hasLower := false
	hasNumber := false
	hasSpecial := false

	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsNumber(r):
			hasNumber = true
		default:
			hasSpecial = true
		}
	}

	if p.RequireUpper && !hasUpper {
		return domainerrors.ErrWeakPassword
	}
	if p.RequireLower && !hasLower {
		return domainerrors.ErrWeakPassword
	}
	if p.RequireNumber && !hasNumber {
		return domainerrors.ErrWeakPassword
	}
	if p.RequireSpecial && !hasSpecial {
		return domainerrors.ErrWeakPassword
	}

	return nil
}
