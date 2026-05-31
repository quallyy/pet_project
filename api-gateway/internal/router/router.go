package router

import (
	"net/http"
	"time"

	"github.com/cargo-platform/api-gateway/internal/config"
	"github.com/cargo-platform/api-gateway/internal/middleware"
	"github.com/cargo-platform/api-gateway/internal/proxy"
	"github.com/cargo-platform/api-gateway/internal/docs"
	"github.com/cargo-platform/api-gateway/pkg/jwt"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/redis/go-redis/v9"
)

func Build(cfg *config.Config, validator *jwt.Validator, rdb *redis.Client) http.Handler {
	r := chi.NewRouter()

	// ── global middleware (every request) ──
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.Timeout(30 * time.Second))

	// rate limit: 120 requests per minute per IP
	limiter := middleware.NewRateLimiter(rdb, 120, time.Minute)
	r.Use(limiter.Middleware)

	// ── auth middleware instance ──
	auth := middleware.NewAuth(validator, rdb)

	// ── health — always public ──
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// ── /auth/* — public, no JWT required ──
	r.Mount("/auth", proxy.New(cfg.AuthServiceURL))

	// ── /orders/* — customers only ──
	r.Group(func(r chi.Router) {
		r.Use(auth.Authenticate)
		r.Use(middleware.RequireRole(jwt.RoleCustomer))
		r.Mount("/orders", proxy.New(cfg.OrderServiceURL))
	})

	// ── /dealer/* — dealers only ──
	r.Group(func(r chi.Router) {
		r.Use(auth.Authenticate)
		r.Use(middleware.RequireRole(jwt.RoleDealer))
		r.Mount("/dealer", proxy.New(cfg.OrderServiceURL))
	})

	// ── /shipments/* — customers and dealers ──
	r.Group(func(r chi.Router) {
		r.Use(auth.Authenticate)
		r.Use(middleware.RequireRole(jwt.RoleCustomer, jwt.RoleDealer))
		r.Mount("/shipments", proxy.New(cfg.ShipmentServiceURL))
	})

	// ── /admin/* — admins only ──
	r.Group(func(r chi.Router) {
		r.Use(auth.Authenticate)
		r.Use(middleware.RequireRole(jwt.RoleAdmin))
		r.Mount("/admin", proxy.New(cfg.AdminServiceURL))
	})

	r.Mount("/docs", docs.Routes())

	return r
}