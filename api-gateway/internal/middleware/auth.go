package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/cargo-platform/api-gateway/pkg/jwt"
	"github.com/redis/go-redis/v9"
)

type contextKey string

const ClaimsKey contextKey = "claims"

type Auth struct {
	validator *jwt.Validator
	rdb       *redis.Client
}

func NewAuth(validator *jwt.Validator, rdb *redis.Client) *Auth {
	return &Auth{validator: validator, rdb: rdb}
}

// Authenticate validates the JWT, checks the blacklist, injects identity headers.
func (a *Auth) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw := extractBearer(r)
		if raw == "" {
			respondUnauthorized(w, "missing token")
			return
		}

		claims, err := a.validator.Parse(raw)
		if err != nil {
			respondUnauthorized(w, "invalid token")
			return
		}

		// one Redis call — check if this token was revoked on logout
		blacklisted, err := a.isBlacklisted(r.Context(), claims.JTI)
		if err != nil {
			// Redis down — fail closed on auth, safer than letting revoked tokens through
			respondUnauthorized(w, "could not verify token")
			return
		}
		if blacklisted {
			respondUnauthorized(w, "token has been revoked")
			return
		}

		// store claims in context for downstream middleware (e.g. RequireRole)
		ctx := context.WithValue(r.Context(), ClaimsKey, claims)
		r = r.WithContext(ctx)

		// inject identity as plain headers — downstream services never parse a JWT
		r.Header.Set("X-User-ID", claims.UserID.String())
		r.Header.Set("X-Role", string(claims.Role))
		r.Header.Set("X-Session-ID", claims.SessionID.String())
		r.Header.Set("X-JTI", claims.JTI)
		if claims.DealerCountry != nil {
			r.Header.Set("X-Dealer-Country", *claims.DealerCountry)
		}

		// strip the bearer token — services don't need it
		r.Header.Del("Authorization")

		next.ServeHTTP(w, r)
	})
}

// RequireRole enforces that the authenticated user has one of the allowed roles.
// Must be chained after Authenticate.
func RequireRole(roles ...jwt.Role) func(http.Handler) http.Handler {
	allowed := make(map[jwt.Role]bool, len(roles))
	for _, role := range roles {
		allowed[role] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(ClaimsKey).(*jwt.Claims)
			if !ok || claims == nil {
				respondUnauthorized(w, "not authenticated")
				return
			}
			if !allowed[claims.Role] {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error":"forbidden"}`))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (a *Auth) isBlacklisted(ctx context.Context, jti string) (bool, error) {
	val, err := a.rdb.Get(ctx, "blacklist:"+jti).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return val == "1", nil
}

func extractBearer(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	return ""
}

func respondUnauthorized(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"error":"` + msg + `"}`))
}