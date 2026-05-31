package postgres

import (
	"context"
	"errors"

	"github.com/cargo-platform/auth-service/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type InviteRepository struct {
	db *pgxpool.Pool
}

func NewInviteRepository(db *pgxpool.Pool) *InviteRepository {
	return &InviteRepository{db: db}
}

func (r *InviteRepository) Create(ctx context.Context, inv *domain.InviteToken) error {
	return r.db.QueryRow(ctx, `
		INSERT INTO invite_tokens (token_hash, email, dealer_country, dealer_name, created_by, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`,
		inv.TokenHash, inv.Email, inv.DealerCountry, inv.DealerName, inv.CreatedBy, inv.ExpiresAt,
	).Scan(&inv.ID, &inv.CreatedAt)
}

func (r *InviteRepository) GetByTokenHash(ctx context.Context, hash string) (*domain.InviteToken, error) {
	inv := &domain.InviteToken{}
	err := r.db.QueryRow(ctx, `
		SELECT id, token_hash, email, dealer_country, dealer_name,
		       created_by, used, expires_at, created_at
		FROM invite_tokens WHERE token_hash = $1
	`, hash).Scan(
		&inv.ID, &inv.TokenHash, &inv.Email, &inv.DealerCountry, &inv.DealerName,
		&inv.CreatedBy, &inv.Used, &inv.ExpiresAt, &inv.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrInvalidInvite
	}
	return inv, err
}

func (r *InviteRepository) MarkUsed(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE invite_tokens SET used = true WHERE id = $1`, id)
	return err
}