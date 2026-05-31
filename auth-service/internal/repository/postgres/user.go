package postgres

import (
	"context"
	"errors"

	"github.com/cargo-platform/auth-service/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	u := &domain.User{}
	err := r.db.QueryRow(ctx, `
		SELECT id, role, phone, email, password_hash, pin_hash,
		       dealer_country, dealer_name, is_active, created_at, updated_at
		FROM users WHERE id = $1
	`, id).Scan(
		&u.ID, &u.Role, &u.Phone, &u.Email, &u.PasswordHash, &u.PinHash,
		&u.DealerCountry, &u.DealerName, &u.IsActive, &u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	return u, err
}

func (r *UserRepository) GetByPhone(ctx context.Context, phone string) (*domain.User, error) {
	u := &domain.User{}
	err := r.db.QueryRow(ctx, `
		SELECT id, role, phone, email, password_hash, pin_hash,
		       dealer_country, dealer_name, is_active, created_at, updated_at
		FROM users WHERE phone = $1
	`, phone).Scan(
		&u.ID, &u.Role, &u.Phone, &u.Email, &u.PasswordHash, &u.PinHash,
		&u.DealerCountry, &u.DealerName, &u.IsActive, &u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	return u, err
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	u := &domain.User{}
	err := r.db.QueryRow(ctx, `
		SELECT id, role, phone, email, password_hash, pin_hash,
		       dealer_country, dealer_name, is_active, created_at, updated_at
		FROM users WHERE email = $1
	`, email).Scan(
		&u.ID, &u.Role, &u.Phone, &u.Email, &u.PasswordHash, &u.PinHash,
		&u.DealerCountry, &u.DealerName, &u.IsActive, &u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	return u, err
}

func (r *UserRepository) Create(ctx context.Context, u *domain.User) error {
	return r.db.QueryRow(ctx, `
		INSERT INTO users (role, phone, email, password_hash, pin_hash, dealer_country, dealer_name)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`,
		u.Role, u.Phone, u.Email, u.PasswordHash, u.PinHash, u.DealerCountry, u.DealerName,
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
}

func (r *UserRepository) SetActive(ctx context.Context, userID uuid.UUID, active bool) error {
	_, err := r.db.Exec(ctx,
		`UPDATE users SET is_active = $1, updated_at = NOW() WHERE id = $2`,
		active, userID,
	)
	return err
}