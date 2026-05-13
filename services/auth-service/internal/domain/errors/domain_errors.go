package domainerrors

import "errors"

var (
	ErrInvalidEmail            = errors.New("invalid email")
	ErrInvalidCredentialStatus = errors.New("invalid credential status")
	ErrInvalidPasswordPolicy   = errors.New("invalid password policy")
	ErrWeakPassword            = errors.New("weak password")
	ErrInvalidTokenTTL         = errors.New("invalid token ttl")
	ErrEmptyID                 = errors.New("id is required")
	ErrEmptyUserID             = errors.New("user id is required")
	ErrEmptyPasswordHash       = errors.New("password hash is required")
	ErrCredentialDisabled      = errors.New("credential is disabled")
	ErrRefreshTokenExpired     = errors.New("refresh token expired")
	ErrRefreshTokenRevoked     = errors.New("refresh token revoked")
	ErrEmptyTokenHash          = errors.New("token hash is required")
)
