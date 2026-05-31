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

	"github.com/cargo-platform/api-gateway/internal/config"
	"github.com/cargo-platform/api-gateway/internal/router"
	"github.com/cargo-platform/api-gateway/pkg/jwt"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := config.Load()

	// ── Redis ──
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

	// ── RSA public key ──
	// gateway only needs the public key — it verifies tokens but never issues them
	publicKey := loadPublicKey(cfg.JWTPublicKeyPath)
	validator := jwt.NewValidator(publicKey)
	slog.Info("JWT validator ready")

	// ── router ──
	h := router.Build(cfg, validator, rdb)

	// ── server ──
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      h,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("gateway started", "port", cfg.Port, "env", cfg.Env)
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
	slog.Info("gateway stopped")
}

func loadPublicKey(path string) *rsa.PublicKey {
	data, err := os.ReadFile(path)
	if err != nil {
		slog.Error("cannot read public key", "path", path, "err", err)
		os.Exit(1)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		slog.Error("invalid PEM block in public key")
		os.Exit(1)
	}

	// try PKIX first (openssl rsa -pubout produces this)
	if key, err := x509.ParsePKIXPublicKey(block.Bytes); err == nil {
		if rsaKey, ok := key.(*rsa.PublicKey); ok {
			return rsaKey
		}
	}

	// fall back to PKCS1
	key, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		slog.Error("cannot parse RSA public key", "err", err)
		os.Exit(1)
	}
	return key
}