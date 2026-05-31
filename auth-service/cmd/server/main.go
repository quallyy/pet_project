package main

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cargo-platform/auth-service/internal/config"
	"github.com/cargo-platform/auth-service/internal/handler"
	pgrepo "github.com/cargo-platform/auth-service/internal/repository/postgres"
	redisrepo "github.com/cargo-platform/auth-service/internal/repository/redis"
	"github.com/cargo-platform/auth-service/internal/service"
	jwtpkg "github.com/cargo-platform/auth-service/pkg/jwt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := config.Load()

	// ── postgres ──
	db, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		slog.Error("cannot connect to postgres", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(context.Background()); err != nil {
		slog.Error("postgres ping failed", "err", err)
		os.Exit(1)
	}
	slog.Info("postgres connected")

	// ── redis ──
	redisOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		slog.Error("invalid redis URL", "err", err)
		os.Exit(1)
	}
	rdb := redis.NewClient(redisOpts)
	// if err := rdb.Ping(context.Background()).Err(); err != nil {
	// 	slog.Error("redis ping failed", "err", err)
	// 	os.Exit(1)
	// }
	slog.Info("redis connected")

	// ── RSA private key ──
	privateKey := loadPrivateKey(cfg.JWTPrivateKeyPath)
	jwtMgr := jwtpkg.NewManager(privateKey)

	// ── repositories ──
	userRepo    := pgrepo.NewUserRepository(db)
	sessionRepo := pgrepo.NewSessionRepository(db)
	inviteRepo  := pgrepo.NewInviteRepository(db)
	auditLogger := pgrepo.NewAuditLogger(db)
	blacklist   := redisrepo.NewTokenBlacklist(rdb)

	// ── service ──
	authSvc := service.NewAuthService(
		userRepo,
		sessionRepo,
		inviteRepo,
		blacklist,
		auditLogger,
		jwtMgr,
	)

	// ── handlers ──
	authHandler := handler.NewAuthHandler(authSvc)

	// ── router ──
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.Mount("/auth", authHandler.Routes())

	// ── server ──
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("auth service started", "port", cfg.Port, "env", cfg.Env)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	<-quit
	slog.Info("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("forced shutdown", "err", err)
	}
	slog.Info("stopped")
}

func loadPrivateKey(path string) *rsa.PrivateKey {
	data, err := os.ReadFile(path)
	if err != nil {
		slog.Error("cannot read private key", "path", path, "err", err)
		os.Exit(1)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		slog.Error("invalid PEM block")
		os.Exit(1)
	}
	// try PKCS8 first, fall back to PKCS1
	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		if rsaKey, ok := key.(*rsa.PrivateKey); ok {
			return rsaKey
		}
	}
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		slog.Error("cannot parse RSA private key", "err", err)
		os.Exit(1)
	}
	return key
}