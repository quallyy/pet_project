package jwt

import (
	"crypto/rsa"
	"errors"
	"time"

	"github.com/cargo-platform/auth-service/internal/domain"
	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const AccessTokenTTL = 15 * time.Minute

type Manager struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

func NewManager(privateKey *rsa.PrivateKey) *Manager {
	return &Manager{
		privateKey: privateKey,
		publicKey:  &privateKey.PublicKey,
	}
}

// PublicKey is handed to the API gateway so it can validate tokens locally
func (m *Manager) PublicKey() *rsa.PublicKey {
	return m.publicKey
}

// internal JWT claims shape
type jwtClaims struct {
	gojwt.RegisteredClaims
	UserID        string  `json:"uid"`
	Role          string  `json:"role"`
	SessionID     string  `json:"sid"`
	DealerCountry *string `json:"dc,omitempty"`
}

func (m *Manager) IssueAccessToken(c domain.Claims) (string, error) {
	now := time.Now()

	var dc *string
	if c.DealerCountry != nil {
		s := string(*c.DealerCountry)
		dc = &s
	}

	claims := jwtClaims{
		RegisteredClaims: gojwt.RegisteredClaims{
			ID:        uuid.New().String(), // jti — used for blacklisting on logout
			IssuedAt:  gojwt.NewNumericDate(now),
			ExpiresAt: gojwt.NewNumericDate(now.Add(AccessTokenTTL)),
			Issuer:    "cargo-auth",
		},
		UserID:        c.UserID.String(),
		Role:          string(c.Role),
		SessionID:     c.SessionID.String(),
		DealerCountry: dc,
	}

	token := gojwt.NewWithClaims(gojwt.SigningMethodRS256, claims)
	return token.SignedString(m.privateKey)
}

type ParsedClaims struct {
	JTI           string
	UserID        uuid.UUID
	Role          domain.Role
	SessionID     uuid.UUID
	DealerCountry *domain.DealerCountry
	ExpiresAt     time.Time
}

func (m *Manager) ParseAccessToken(raw string) (*ParsedClaims, error) {
	token, err := gojwt.ParseWithClaims(raw, &jwtClaims{}, func(t *gojwt.Token) (any, error) {
		if _, ok := t.Method.(*gojwt.SigningMethodRSA); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return m.publicKey, nil
	})
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	c, ok := token.Claims.(*jwtClaims)
	if !ok || !token.Valid {
		return nil, domain.ErrInvalidToken
	}

	userID, err := uuid.Parse(c.UserID)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}
	sessionID, err := uuid.Parse(c.SessionID)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	out := &ParsedClaims{
		JTI:       c.ID,
		UserID:    userID,
		Role:      domain.Role(c.Role),
		SessionID: sessionID,
		ExpiresAt: c.ExpiresAt.Time,
	}
	if c.DealerCountry != nil {
		dc := domain.DealerCountry(*c.DealerCountry)
		out.DealerCountry = &dc
	}
	return out, nil
}