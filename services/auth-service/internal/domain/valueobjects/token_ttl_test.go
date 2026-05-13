package valueobjects

import (
	"testing"
	"time"
)

func TestNewTokenTTL(t *testing.T) {
	_, err := NewTokenTTL(0, time.Hour)
	if err == nil {
		t.Fatal("expected invalid ttl error when access ttl <= 0")
	}

	_, err = NewTokenTTL(time.Minute, time.Minute)
	if err == nil {
		t.Fatal("expected invalid ttl error when refresh <= access")
	}

	ttl, err := NewTokenTTL(15*time.Minute, 24*time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ttl.AccessSeconds() != int64((15 * time.Minute).Seconds()) {
		t.Fatalf("unexpected access seconds: %d", ttl.AccessSeconds())
	}
}
