package jwt

import (
	"crypto/rsa"
	"errors"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Role string

const (
	RoleCustomer Role = "customer"
	RoleDealer   Role = "dealer"
	RoleAdmin    Role = "admin"
)

type Claims struct {
	JTI           string
	UserID        uuid.UUID
	Role          Role
	SessionID     uuid.UUID
	DealerCountry *string // "CN" | "TR" | "RU" — only set for dealers
	ExpiresAt     time.Time
}

type Validator struct {
	publicKey *rsa.PublicKey
}

func NewValidator(publicKey *rsa.PublicKey) *Validator {
	return &Validator{publicKey: publicKey}
}

type rawClaims struct {
	gojwt.RegisteredClaims
	UserID        string  `json:"uid"`
	Role          string  `json:"role"`
	SessionID     string  `json:"sid"`
	DealerCountry *string `json:"dc,omitempty"`
}

// Parse validates the token signature and expiry, returns typed claims.
// Purely local — no network call, no Redis, just crypto.
func (v *Validator) Parse(tokenStr string) (*Claims, error) {
	token, err := gojwt.ParseWithClaims(tokenStr, &rawClaims{}, func(t *gojwt.Token) (any, error) {
		if _, ok := t.Method.(*gojwt.SigningMethodRSA); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return v.publicKey, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	c, ok := token.Claims.(*rawClaims)
	if !ok {
		return nil, errors.New("malformed claims")
	}

	userID, err := uuid.Parse(c.UserID)
	if err != nil {
		return nil, errors.New("invalid user_id in token")
	}
	sessionID, err := uuid.Parse(c.SessionID)
	if err != nil {
		return nil, errors.New("invalid session_id in token")
	}

	return &Claims{
		JTI:           c.ID,
		UserID:        userID,
		Role:          Role(c.Role),
		SessionID:     sessionID,
		DealerCountry: c.DealerCountry,
		ExpiresAt:     c.ExpiresAt.Time,
	}, nil
}