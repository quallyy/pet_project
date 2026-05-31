package postgres

import (
	"context"
	"errors"

	"github.com/cargo-platform/auth-service/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SessionRepository struct {
	db *pgxpool.Pool
}

func NewSessionRepository(db *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(ctx context.Context, s *domain.Session) error {
	return r.db.QueryRow(ctx, `
		INSERT INTO sessions (id, user_id, refresh_token, device_info, ip_address, expires_at)
		VALUES ($1, $2, $3, $4, $5::inet, $6)
		RETURNING created_at, last_used_at
	`,
		s.ID, s.UserID, s.RefreshToken, s.DeviceInfo, s.IPAddress, s.ExpiresAt,
	).Scan(&s.CreatedAt, &s.LastUsedAt)
}

func (r *SessionRepository) GetByRefreshToken(ctx context.Context, tokenHash string) (*domain.Session, error) {
	s := &domain.Session{}
	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, refresh_token, device_info, ip_address,
		       expires_at, created_at, last_used_at
		FROM sessions WHERE refresh_token = $1
	`, tokenHash).Scan(
		&s.ID, &s.UserID, &s.RefreshToken, &s.DeviceInfo, &s.IPAddress,
		&s.ExpiresAt, &s.CreatedAt, &s.LastUsedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrSessionNotFound
	}
	return s, err
}

func (r *SessionRepository) DeleteByID(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM sessions WHERE id = $1`, id)
	return err
}

func (r *SessionRepository) DeleteAllForUser(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM sessions WHERE user_id = $1`, userID)
	return err
}

func (r *SessionRepository) CountForUser(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM sessions WHERE user_id = $1`, userID,
	).Scan(&count)
	return count, err
}

func (r *SessionRepository) DeleteOldestForUser(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		DELETE FROM sessions
		WHERE id = (
			SELECT id FROM sessions
			WHERE user_id = $1
			ORDER BY last_used_at ASC
			LIMIT 1
		)
	`, userID)
	return err
}