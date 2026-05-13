package valueobjects

import "testing"

func TestCredentialStatus(t *testing.T) {
	if !CredentialStatusActive.IsValid() {
		t.Fatal("active should be valid")
	}
	if !CredentialStatusDisabled.IsValid() {
		t.Fatal("disabled should be valid")
	}
	if CredentialStatus("inactive").IsValid() {
		t.Fatal("inactive string should not be valid")
	}

	if CredentialStatusInactive != CredentialStatusDisabled {
		t.Fatal("expected inactive alias to map to disabled")
	}

	if _, err := NewCredentialStatus("active"); err != nil {
		t.Fatalf("unexpected error for active: %v", err)
	}
	if _, err := NewCredentialStatus("disabled"); err != nil {
		t.Fatalf("unexpected error for disabled: %v", err)
	}
	if _, err := NewCredentialStatus("inactive"); err == nil {
		t.Fatal("expected error for inactive")
	}
}
