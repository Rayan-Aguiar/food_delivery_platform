package valueobjects

import "time"

type TokenClaims struct {
	Subject   string
	IssuedAt  time.Time
	ExpiresAt time.Time
}

func (c TokenClaims) IsExpired(now time.Time) bool {
	return now.After(c.ExpiresAt)
}
