package mongo

import (
	"testing"
	"time"
)

func TestMapRefreshDocToEntity_AllFields(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	revokedAt := now.Add(2 * time.Hour)
	rotatedFromID := "parent-session-id"

	doc := refreshSessionDoc{
		ID:                   "session-1",
		UserID:               "user-1",
		TokenHash:            "hash-abc",
		ExpiresAt:            now.Add(7 * 24 * time.Hour),
		RevokedAt:            &revokedAt,
		RotatedFromSessionID: &rotatedFromID,
		UserAgent:            "Mozilla/5.0",
		IPAddress:            "192.168.0.1",
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	entity := mapRefreshDocToEntity(doc)

	if entity.ID != doc.ID {
		t.Errorf("ID = %q, want %q", entity.ID, doc.ID)
	}
	if entity.UserID != doc.UserID {
		t.Errorf("UserID = %q, want %q", entity.UserID, doc.UserID)
	}
	if entity.TokenHash != doc.TokenHash {
		t.Errorf("TokenHash = %q, want %q", entity.TokenHash, doc.TokenHash)
	}
	if !entity.ExpiresAt.Equal(doc.ExpiresAt) {
		t.Errorf("ExpiresAt = %v, want %v", entity.ExpiresAt, doc.ExpiresAt)
	}
	if entity.RevokedAt == nil || !entity.RevokedAt.Equal(revokedAt) {
		t.Errorf("RevokedAt = %v, want %v", entity.RevokedAt, revokedAt)
	}
	if entity.RotatedFromSessionID == nil || *entity.RotatedFromSessionID != rotatedFromID {
		t.Errorf("RotatedFromSessionID = %v, want %q", entity.RotatedFromSessionID, rotatedFromID)
	}
	if entity.UserAgent != doc.UserAgent {
		t.Errorf("UserAgent = %q, want %q", entity.UserAgent, doc.UserAgent)
	}
	if entity.IPAddress != doc.IPAddress {
		t.Errorf("IPAddress = %q, want %q", entity.IPAddress, doc.IPAddress)
	}
	if !entity.CreatedAt.Equal(doc.CreatedAt) {
		t.Errorf("CreatedAt = %v, want %v", entity.CreatedAt, doc.CreatedAt)
	}
}

func TestMapRefreshDocToEntity_NilOptionalFields(t *testing.T) {
	now := time.Now()
	doc := refreshSessionDoc{
		ID:                   "session-2",
		UserID:               "user-2",
		TokenHash:            "hash-def",
		ExpiresAt:            now.Add(24 * time.Hour),
		RevokedAt:            nil,
		RotatedFromSessionID: nil,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	entity := mapRefreshDocToEntity(doc)

	if entity.RevokedAt != nil {
		t.Errorf("RevokedAt deve ser nil, got %v", entity.RevokedAt)
	}
	if entity.RotatedFromSessionID != nil {
		t.Errorf("RotatedFromSessionID deve ser nil, got %v", entity.RotatedFromSessionID)
	}
}
