package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/cargo-platform/auth-service/internal/domain"
	"github.com/cargo-platform/auth-service/pkg/hash"
	"github.com/cargo-platform/auth-service/pkg/jwt"
	"github.com/google/uuid"
)

const (
	maxSessions     = 3
	refreshTokenTTL = 30 * 24 * time.Hour
	inviteTTL       = 48 * time.Hour
)

type AuthService struct {
	users     domain.UserRepository
	sessions  domain.SessionRepository
	invites   domain.InviteRepository
	blacklist domain.TokenBlacklist
	audit     domain.AuditLogger
	jwt       *jwt.Manager
}

func NewAuthService(
	users domain.UserRepository,
	sessions domain.SessionRepository,
	invites domain.InviteRepository,
	blacklist domain.TokenBlacklist,
	audit domain.AuditLogger,
	jwtMgr *jwt.Manager,
) *AuthService {
	return &AuthService{
		users:     users,
		sessions:  sessions,
		invites:   invites,
		blacklist: blacklist,
		audit:     audit,
		jwt:       jwtMgr,
	}
}

// ── DEALER REGISTRATION ───────────────────────────────────────────────────────

// RegisterDealer validates an invite token and creates the dealer account
func (s *AuthService) RegisterDealer(ctx context.Context, rawToken, password, deviceInfo, ip string) (*domain.TokenPair, error) {
	tokenHash := hash.Token(rawToken)

	invite, err := s.invites.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, domain.ErrInvalidInvite
	}
	if invite.Used || time.Now().After(invite.ExpiresAt) {
		return nil, domain.ErrInvalidInvite
	}

	// guard against double-registration on the same email
	if _, err := s.users.GetByEmail(ctx, invite.Email); err == nil {
		return nil, domain.ErrEmailTaken
	}

	pwHash, err := hash.Password(password)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Role:          domain.RoleDealer,
		Email:         &invite.Email,
		PasswordHash:  &pwHash,
		DealerCountry: &invite.DealerCountry,
		DealerName:    &invite.DealerName,
	}
	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}

	if err := s.invites.MarkUsed(ctx, invite.ID); err != nil {
		slog.Error("failed to mark invite used", "invite_id", invite.ID, "err", err)
	}

	s.audit.Log(ctx, "dealer.registered", &user.ID, ip, map[string]any{
		"email":   invite.Email,
		"country": invite.DealerCountry,
	})

	return s.createSession(ctx, user, deviceInfo, ip)
}

// ── LOGIN ─────────────────────────────────────────────────────────────────────

// Login authenticates dealers and admins via email + password
func (s *AuthService) Login(ctx context.Context, email, password, deviceInfo, ip string) (*domain.TokenPair, error) {
	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		// dummy check keeps response time constant — prevents user enumeration
		_ = hash.CheckPassword("dummy", "$2a$12$aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
		s.audit.Log(ctx, "login.failed", nil, ip, map[string]any{"email": email, "reason": "not_found"})
		return nil, domain.ErrInvalidCredentials
	}

	if user.PasswordHash == nil || !hash.CheckPassword(password, *user.PasswordHash) {
		s.audit.Log(ctx, "login.failed", &user.ID, ip, map[string]any{"reason": "wrong_password"})
		return nil, domain.ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, domain.ErrUserInactive
	}

	s.audit.Log(ctx, "login.success", &user.ID, ip, nil)
	return s.createSession(ctx, user, deviceInfo, ip)
}

// ── TOKEN LIFECYCLE ───────────────────────────────────────────────────────────

// Refresh rotates the refresh token and issues a new access token
func (s *AuthService) Refresh(ctx context.Context, rawRefreshToken, deviceInfo, ip string) (*domain.TokenPair, error) {
	tokenHash := hash.Token(rawRefreshToken)

	session, err := s.sessions.GetByRefreshToken(ctx, tokenHash)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}
	if time.Now().After(session.ExpiresAt) {
		_ = s.sessions.DeleteByID(ctx, session.ID)
		return nil, domain.ErrInvalidToken
	}

	user, err := s.users.GetByID(ctx, session.UserID)
	if err != nil || !user.IsActive {
		return nil, domain.ErrInvalidToken
	}

	// rotate: old session out, new session in
	_ = s.sessions.DeleteByID(ctx, session.ID)
	return s.createSession(ctx, user, deviceInfo, ip)
}

// Logout revokes the current session and blacklists the JWT jti
func (s *AuthService) Logout(ctx context.Context, sessionID uuid.UUID, jti string) error {
	_ = s.sessions.DeleteByID(ctx, sessionID)
	// blacklist for the remaining JWT lifetime (access tokens live 15 min)
	return s.blacklist.Add(ctx, jti, jwt.AccessTokenTTL)
}

// LogoutAll terminates every session — used when a phone is stolen
func (s *AuthService) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	return s.sessions.DeleteAllForUser(ctx, userID)
}

// ── INVITE MANAGEMENT ────────────────────────────────────────────────────────

// CreateInvite generates a dealer invite link — called by admins only
func (s *AuthService) CreateInvite(
	ctx context.Context,
	adminID uuid.UUID,
	email string,
	country domain.DealerCountry,
	dealerName string,
) (string, error) {
	// two UUIDs concatenated = 72 chars of entropy — plenty
	rawToken := uuid.New().String() + uuid.New().String()

	invite := &domain.InviteToken{
		TokenHash:     hash.Token(rawToken),
		Email:         email,
		DealerCountry: country,
		DealerName:    dealerName,
		CreatedBy:     adminID,
		ExpiresAt:     time.Now().Add(inviteTTL),
	}
	if err := s.invites.Create(ctx, invite); err != nil {
		return "", err
	}

	s.audit.Log(ctx, "invite.created", &adminID, "", map[string]any{
		"email":   email,
		"country": country,
	})

	return rawToken, nil
}

// ── INTERNAL ─────────────────────────────────────────────────────────────────

func (s *AuthService) createSession(ctx context.Context, user *domain.User, deviceInfo, ip string) (*domain.TokenPair, error) {
	// enforce session cap — evict oldest if at limit
	count, err := s.sessions.CountForUser(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	if count >= maxSessions {
		if err := s.sessions.DeleteOldestForUser(ctx, user.ID); err != nil {
			return nil, err
		}
	}

	sessionID := uuid.New()
	rawRefresh := uuid.New().String()

	session := &domain.Session{
		ID:           sessionID,
		UserID:       user.ID,
		RefreshToken: hash.Token(rawRefresh),
		DeviceInfo:   deviceInfo,
		IPAddress:    ip,
		ExpiresAt:    time.Now().Add(refreshTokenTTL),
	}
	if err := s.sessions.Create(ctx, session); err != nil {
		return nil, err
	}

	accessToken, err := s.jwt.IssueAccessToken(domain.Claims{
		UserID:        user.ID,
		Role:          user.Role,
		SessionID:     sessionID,
		DealerCountry: user.DealerCountry,
	})
	if err != nil {
		return nil, err
	}

	return &domain.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		ExpiresIn:    int(jwt.AccessTokenTTL.Seconds()),
	}, nil
}