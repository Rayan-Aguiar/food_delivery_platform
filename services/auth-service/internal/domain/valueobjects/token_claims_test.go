package valueobjects

import (
	"testing"
	"time"
)

func TestTokenClaimsIsExpired(t *testing.T) {
	now := time.Now()
	c := TokenClaims{Subject: "u1", IssuedAt: now, ExpiresAt: now.Add(1 * time.Minute)}
	if c.IsExpired(now) {
		t.Fatal("claims should not be expired at now")
	}
	if !c.IsExpired(now.Add(2 * time.Minute)) {
		t.Fatal("claims should be expired after exp")
	}
}
