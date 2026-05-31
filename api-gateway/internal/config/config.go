package config

import "os"

type Config struct {
	Port              string
	RedisURL          string
	JWTPublicKeyPath  string
	Env               string

	// upstream service addresses
	AuthServiceURL     string
	OrderServiceURL    string
	ShipmentServiceURL string
	AdminServiceURL    string
}

func Load() *Config {
	return &Config{
		Port:               getEnv("PORT", "8080"),
		RedisURL:           getEnv("REDIS_URL", "redis://localhost:6379"),
		JWTPublicKeyPath:   getEnv("JWT_PUBLIC_KEY_PATH", "./keys/public.pem"),
		Env:                getEnv("ENV", "development"),
		AuthServiceURL:     getEnv("AUTH_SERVICE_URL", "http://localhost:8081"),
		OrderServiceURL:    getEnv("ORDER_SERVICE_URL", "http://localhost:8082"),
		ShipmentServiceURL: getEnv("SHIPMENT_SERVICE_URL", "http://localhost:8083"),
		AdminServiceURL:    getEnv("ADMIN_SERVICE_URL", "http://localhost:8084"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}