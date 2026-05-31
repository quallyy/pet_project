package domain

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserInactive       = errors.New("account is inactive")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrSessionNotFound    = errors.New("session not found")

	ErrInvalidInvite = errors.New("invite token is invalid or expired")
	ErrInviteUsed    = errors.New("invite token already used")
	ErrEmailTaken    = errors.New("email already registered")

	ErrMaxSessions = errors.New("maximum active sessions reached")
	ErrForbidden   = errors.New("insufficient permissions")
)