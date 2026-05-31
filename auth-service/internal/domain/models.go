package domain

import (
	"time"

	"github.com/google/uuid"
)

type Role string
type DealerCountry string

const (
	RoleCustomer Role = "customer"
	RoleDealer   Role = "dealer"
	RoleAdmin    Role = "admin"

	CountryCN DealerCountry = "CN"
	CountryTR DealerCountry = "TR"
	CountryRU DealerCountry = "RU"
)

type User struct {
	ID            uuid.UUID      `db:"id"`
	Role          Role           `db:"role"`
	Phone         *string        `db:"phone"`          // customers
	Email         *string        `db:"email"`          // dealers + admins
	PasswordHash  *string        `db:"password_hash"`  // dealers + admins
	PinHash       *string        `db:"pin_hash"`       // optional, customers
	DealerCountry *DealerCountry `db:"dealer_country"`
	DealerName    *string        `db:"dealer_name"`
	IsActive      bool           `db:"is_active"`
	CreatedAt     time.Time      `db:"created_at"`
	UpdatedAt     time.Time      `db:"updated_at"`
}

type Session struct {
	ID           uuid.UUID `db:"id"`
	UserID       uuid.UUID `db:"user_id"`
	RefreshToken string    `db:"refresh_token"` // stored hashed
	DeviceInfo   string    `db:"device_info"`
	IPAddress    string    `db:"ip_address"`
	ExpiresAt    time.Time `db:"expires_at"`
	CreatedAt    time.Time `db:"created_at"`
	LastUsedAt   time.Time `db:"last_used_at"`
}

type InviteToken struct {
	ID            uuid.UUID     `db:"id"`
	TokenHash     string        `db:"token_hash"`
	Email         string        `db:"email"`
	DealerCountry DealerCountry `db:"dealer_country"`
	DealerName    string        `db:"dealer_name"`
	CreatedBy     uuid.UUID     `db:"created_by"`
	Used          bool          `db:"used"`
	ExpiresAt     time.Time     `db:"expires_at"`
	CreatedAt     time.Time     `db:"created_at"`
}

// TokenPair is returned to clients on every successful auth
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"` // seconds
}

// Claims is embedded inside the JWT — kept lean on purpose
type Claims struct {
	UserID        uuid.UUID
	Role          Role
	SessionID     uuid.UUID
	DealerCountry *DealerCountry
}