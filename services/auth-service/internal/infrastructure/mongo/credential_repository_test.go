package mongo

import (
	"testing"
	"time"

	"food_delivery_platform/services/auth-service/internal/domain/valueobjects"
)

func TestMapCredentialDocToEntity_Success(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	lastLogin := now.Add(-1 * time.Hour)

	doc := credentialDoc{
		ID:                  "cred-1",
		UserID:              "user-1",
		Email:               "user@example.com",
		PasswordHash:        "hash123",
		Status:              string(valueobjects.CredentialStatusActive),
		FailedLoginAttempts: 2,
		LastLoginAt:         &lastLogin,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	entity, err := mapCredentialDocToEntity(doc)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if entity.ID != doc.ID {
		t.Errorf("ID = %q, want %q", entity.ID, doc.ID)
	}
	if entity.UserID != doc.UserID {
		t.Errorf("UserID = %q, want %q", entity.UserID, doc.UserID)
	}
	if entity.Email.String() != doc.Email {
		t.Errorf("Email = %q, want %q", entity.Email.String(), doc.Email)
	}
	if entity.PasswordHash != doc.PasswordHash {
		t.Errorf("PasswordHash = %q, want %q", entity.PasswordHash, doc.PasswordHash)
	}
	if entity.Status != valueobjects.CredentialStatusActive {
		t.Errorf("Status = %q, want %q", entity.Status, valueobjects.CredentialStatusActive)
	}
	if entity.FailedLoginAttempts != doc.FailedLoginAttempts {
		t.Errorf("FailedLoginAttempts = %d, want %d", entity.FailedLoginAttempts, doc.FailedLoginAttempts)
	}
	if entity.LastLoginAt == nil || !entity.LastLoginAt.Equal(lastLogin) {
		t.Errorf("LastLoginAt = %v, want %v", entity.LastLoginAt, lastLogin)
	}
	if !entity.CreatedAt.Equal(doc.CreatedAt) {
		t.Errorf("CreatedAt = %v, want %v", entity.CreatedAt, doc.CreatedAt)
	}
}

func TestMapCredentialDocToEntity_NilLastLoginAt(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	doc := credentialDoc{
		ID:           "cred-2",
		UserID:       "user-2",
		Email:        "other@example.com",
		PasswordHash: "hash456",
		Status:       string(valueobjects.CredentialStatusActive),
		LastLoginAt:  nil,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	entity, err := mapCredentialDocToEntity(doc)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if entity.LastLoginAt != nil {
		t.Errorf("LastLoginAt deve ser nil, got %v", entity.LastLoginAt)
	}
}

func TestMapCredentialDocToEntity_InvalidEmail(t *testing.T) {
	now := time.Now()
	doc := credentialDoc{
		ID:           "cred-x",
		UserID:       "user-x",
		Email:        "not-a-valid-email",
		PasswordHash: "hash",
		Status:       string(valueobjects.CredentialStatusActive),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	_, err := mapCredentialDocToEntity(doc)
	if err == nil {
		t.Error("esperava erro para email inválido, got nil")
	}
}

func TestMapCredentialDocToEntity_InvalidStatus(t *testing.T) {
	now := time.Now()
	doc := credentialDoc{
		ID:           "cred-x",
		UserID:       "user-x",
		Email:        "user@example.com",
		PasswordHash: "hash",
		Status:       "unknown_status",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	_, err := mapCredentialDocToEntity(doc)
	if err == nil {
		t.Error("esperava erro para status inválido, got nil")
	}
}
