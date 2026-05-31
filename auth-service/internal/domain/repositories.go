package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type UserRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByPhone(ctx context.Context, phone string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Create(ctx context.Context, user *User) error
	SetActive(ctx context.Context, userID uuid.UUID, active bool) error
}

type SessionRepository interface {
	Create(ctx context.Context, session *Session) error
	GetByRefreshToken(ctx context.Context, tokenHash string) (*Session, error)
	DeleteByID(ctx context.Context, id uuid.UUID) error
	DeleteAllForUser(ctx context.Context, userID uuid.UUID) error
	CountForUser(ctx context.Context, userID uuid.UUID) (int, error)
	DeleteOldestForUser(ctx context.Context, userID uuid.UUID) error
}

type InviteRepository interface {
	Create(ctx context.Context, invite *InviteToken) error
	GetByTokenHash(ctx context.Context, hash string) (*InviteToken, error)
	MarkUsed(ctx context.Context, id uuid.UUID) error
}

type TokenBlacklist interface {
	Add(ctx context.Context, jti string, ttl time.Duration) error
	IsBlacklisted(ctx context.Context, jti string) (bool, error)
}

type AuditLogger interface {
	Log(ctx context.Context, event string, userID *uuid.UUID, ip string, meta map[string]any)
}