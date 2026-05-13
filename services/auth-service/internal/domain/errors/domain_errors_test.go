package domainerrors

import "testing"

func TestDomainErrorsDefined(t *testing.T) {
	errs := []error{
		ErrInvalidEmail,
		ErrInvalidCredentialStatus,
		ErrInvalidPasswordPolicy,
		ErrWeakPassword,
		ErrInvalidTokenTTL,
		ErrEmptyID,
		ErrEmptyUserID,
		ErrEmptyPasswordHash,
		ErrCredentialDisabled,
		ErrRefreshTokenExpired,
		ErrRefreshTokenRevoked,
		ErrEmptyTokenHash,
	}

	seen := map[string]bool{}
	for _, err := range errs {
		if err == nil {
			t.Fatal("expected non-nil domain error")
		}
		if err.Error() == "" {
			t.Fatal("expected non-empty error message")
		}
		if seen[err.Error()] {
			t.Fatalf("duplicated domain error message: %s", err.Error())
		}
		seen[err.Error()] = true
	}
}
