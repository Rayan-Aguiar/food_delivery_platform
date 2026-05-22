package mongo

import "time"

type credentialDoc struct {
	ID                  string     `bson:"_id"`
	UserID              string     `bson:"user_id"`
	Email               string     `bson:"email"`
	PasswordHash        string     `bson:"password_hash"`
	Status              string     `bson:"status"`
	FailedLoginAttempts int        `bson:"failed_login_attempts"`
	LastLoginAt         *time.Time `bson:"last_login_at,omitempty"`
	CreatedAt           time.Time  `bson:"created_at"`
	UpdatedAt           time.Time  `bson:"updated_at"`
}

type refreshSessionDoc struct {
	ID                   string     `bson:"_id"`
	UserID               string     `bson:"user_id"`
	TokenHash            string     `bson:"token_hash"`
	ExpiresAt            time.Time  `bson:"expires_at"`
	RevokedAt            *time.Time `bson:"revoked_at,omitempty"`
	RotatedFromSessionID *string    `bson:"rotated_from_session_id,omitempty"`
	UserAgent            string     `bson:"user_agent,omitempty"`
	IPAddress            string     `bson:"ip_address,omitempty"`
	CreatedAt            time.Time  `bson:"created_at"`
	UpdatedAt            time.Time  `bson:"updated_at"`
}

type outboxMessageDoc struct {
	ID         string            `bson:"_id"`
	Exchange   string            `bson:"exchange"`
	RoutingKey string            `bson:"routing_key"`
	Body       []byte            `bson:"body"`
	Headers    map[string]string `bson:"headers"`
	CreatedAt  time.Time         `bson:"created_at"`
	LastError  string            `bson:"last_error,omitempty"`
}
