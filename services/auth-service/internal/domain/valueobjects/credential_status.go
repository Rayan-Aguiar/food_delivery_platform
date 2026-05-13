package valueobjects

import domainerrors "food_delivery_platform/services/auth-service/internal/domain/errors"

type CredentialStatus string

const (
	CredentialStatusActive   CredentialStatus = "active"
	CredentialStatusDisabled CredentialStatus = "disabled"
	// Backward compatibility alias while codebase migrates nomenclature.
	CredentialStatusInactive CredentialStatus = CredentialStatusDisabled
)

func NewCredentialStatus(raw string) (CredentialStatus, error) {
	s := CredentialStatus(raw)
	if !s.IsValid() {
		return "", domainerrors.ErrInvalidCredentialStatus
	}
	return s, nil
}

func (s CredentialStatus) IsValid () bool {
	return s == CredentialStatusActive || s == CredentialStatusDisabled
}
