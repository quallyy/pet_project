package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/cargo-platform/auth-service/internal/domain"
	"github.com/cargo-platform/auth-service/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type AuthHandler struct {
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Routes() chi.Router {
	r := chi.NewRouter()

	// public — no token required
	r.Post("/login", h.Login)
	r.Post("/refresh", h.Refresh)
	r.Post("/dealer/register", h.RegisterDealer)

	// requires a valid access token (gateway validates before forwarding)
	r.Post("/logout", h.Logout)
	r.Post("/logout-all", h.LogoutAll)

	// admin only (gateway enforces role before forwarding)
	r.Post("/admin/invite", h.CreateInvite)

	return r
}

// POST /auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !decode(w, r, &req) {
		return
	}
	if req.Email == "" || req.Password == "" {
		respondErr(w, http.StatusBadRequest, "email and password are required")
		return
	}

	pair, err := h.svc.Login(r.Context(), req.Email, req.Password, r.UserAgent(), realIP(r))
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidCredentials), errors.Is(err, domain.ErrUserInactive):
			respondErr(w, http.StatusUnauthorized, "invalid credentials")
		default:
			respondErr(w, http.StatusInternalServerError, "login failed")
		}
		return
	}

	respond(w, http.StatusOK, pair)
}

// POST /auth/refresh
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if !decode(w, r, &req) {
		return
	}
	if req.RefreshToken == "" {
		respondErr(w, http.StatusBadRequest, "refresh_token is required")
		return
	}

	pair, err := h.svc.Refresh(r.Context(), req.RefreshToken, r.UserAgent(), realIP(r))
	if err != nil {
		respondErr(w, http.StatusUnauthorized, "invalid or expired refresh token")
		return
	}

	respond(w, http.StatusOK, pair)
}

// POST /auth/dealer/register
func (h *AuthHandler) RegisterDealer(w http.ResponseWriter, r *http.Request) {
	var req struct {
		InviteToken string `json:"invite_token"`
		Password    string `json:"password"`
	}
	if !decode(w, r, &req) {
		return
	}
	if req.InviteToken == "" || req.Password == "" {
		respondErr(w, http.StatusBadRequest, "invite_token and password are required")
		return
	}

	pair, err := h.svc.RegisterDealer(r.Context(), req.InviteToken, req.Password, r.UserAgent(), realIP(r))
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidInvite), errors.Is(err, domain.ErrInviteUsed):
			respondErr(w, http.StatusBadRequest, "invalid or expired invite")
		case errors.Is(err, domain.ErrEmailTaken):
			respondErr(w, http.StatusConflict, "email already registered")
		default:
			respondErr(w, http.StatusInternalServerError, "registration failed")
		}
		return
	}

	respond(w, http.StatusCreated, pair)
}

// POST /auth/logout
// Gateway injects X-User-ID, X-Session-ID, X-JTI from the validated JWT
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	sessionID, err := uuid.Parse(r.Header.Get("X-Session-ID"))
	jti := r.Header.Get("X-JTI")
	if err != nil || jti == "" {
		respondErr(w, http.StatusUnauthorized, "missing session")
		return
	}

	if err := h.svc.Logout(r.Context(), sessionID, jti); err != nil {
		respondErr(w, http.StatusInternalServerError, "logout failed")
		return
	}

	respond(w, http.StatusOK, map[string]string{"message": "logged out"})
}

// POST /auth/logout-all — kills all sessions (stolen phone)
func (h *AuthHandler) LogoutAll(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(r.Header.Get("X-User-ID"))
	if err != nil {
		respondErr(w, http.StatusUnauthorized, "missing user")
		return
	}

	if err := h.svc.LogoutAll(r.Context(), userID); err != nil {
		respondErr(w, http.StatusInternalServerError, "failed")
		return
	}

	respond(w, http.StatusOK, map[string]string{"message": "all sessions terminated"})
}

// POST /auth/admin/invite
func (h *AuthHandler) CreateInvite(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email         string               `json:"email"`
		DealerCountry domain.DealerCountry `json:"dealer_country"` // "CN" | "TR" | "RU"
		DealerName    string               `json:"dealer_name"`
	}
	if !decode(w, r, &req) {
		return
	}
	if req.Email == "" || req.DealerCountry == "" || req.DealerName == "" {
		respondErr(w, http.StatusBadRequest, "email, dealer_country and dealer_name are required")
		return
	}

	adminID, err := uuid.Parse(r.Header.Get("X-User-ID"))
	if err != nil {
		respondErr(w, http.StatusUnauthorized, "missing admin identity")
		return
	}

	rawToken, err := h.svc.CreateInvite(r.Context(), adminID, req.Email, req.DealerCountry, req.DealerName)
	if err != nil {
		respondErr(w, http.StatusInternalServerError, "failed to create invite")
		return
	}

	respond(w, http.StatusCreated, map[string]any{
		"invite_url":   "https://cargo.tm/dealer/register?token=" + rawToken,
		"expires_in":   (48 * time.Hour).String(),
	})
}

// ── helpers ───────────────────────────────────────────────────────────────────

func decode(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		respondErr(w, http.StatusBadRequest, "invalid request body")
		return false
	}
	return true
}

func respond(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func respondErr(w http.ResponseWriter, status int, msg string) {
	respond(w, status, map[string]string{"error": msg})
}

func realIP(r *http.Request) string {
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return strings.Split(ip, ",")[0]
	}
	return r.RemoteAddr
}