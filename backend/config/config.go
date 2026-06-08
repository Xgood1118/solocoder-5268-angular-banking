package config

import "os"

type Config struct {
	Port         string
	DatabaseURL  string
	RedisURL     string
	JWTSecret    string
	Pepper       string
	AuditHMACKey string
}

func Load() *Config {
	return &Config{
		Port:         getEnv("PORT", "8080"),
		DatabaseURL:  getEnv("DATABASE_URL", "banking.db"),
		RedisURL:     getEnv("REDIS_URL", "memory"),
		JWTSecret:    getEnv("JWT_SECRET", "your-jwt-secret-key-change-in-production"),
		Pepper:       getEnv("PEPPER", "banking-pepper-key-2024"),
		AuditHMACKey: getEnv("AUDIT_HMAC_KEY", "audit-hmac-secret-key"),
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
